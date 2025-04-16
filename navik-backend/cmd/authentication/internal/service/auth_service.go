package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"authentication/internal/domain"
	"authentication/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("token has expired")
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Enabled  bool
}

type JWTConfig struct {
	AccessTokenSecret   string
	RefreshTokenSecret  string
	AccessTokenExpiry   time.Duration
	RefreshTokenExpiry  time.Duration
	Issuer              string
	AccessTokenCacheTTL time.Duration
}

type AuthService struct {
	userRepo    *repository.UserRepository
	jwtConfig   JWTConfig
	redisClient *redis.Client
	useCache    bool
}

func NewAuthService(userRepo *repository.UserRepository, jwtConfig JWTConfig, redisConfig RedisConfig) *AuthService {

	var redisClient *redis.Client
	useCache := false

	if redisConfig.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConfig.Addr,
			Password: redisConfig.Password,
			DB:       redisConfig.DB,
		})

		// Test connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if _, err := redisClient.Ping(ctx).Result(); err == nil {
			useCache = true
		} else {
			fmt.Printf("Warning: Redis connection failed: %v, continuing without caching\n", err)
		}
		fmt.Printf("Redis connection established: %s\n", redisConfig.Addr)
	}

	return &AuthService{
		userRepo:    userRepo,
		jwtConfig:   jwtConfig,
		redisClient: redisClient,
		useCache:    useCache,
	}
}

func (s *AuthService) RegisterCustomer(ctx context.Context, reg domain.Registration) (*domain.AuthResponse, error) {
	if reg.UserType != domain.UserTypeCustomer {
		return nil, repository.ErrInvalidUserType
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user with customer-specific fields
	user := domain.User{
		Email:     reg.Email,
		Password:  string(hashedPassword),
		UserType:  reg.UserType,
		Phone:     reg.Phone,
		FirstName: reg.FirstName,
		LastName:  reg.LastName,
		IsActive:  true,
	}

	// Create user in database
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	// Generate tokens
	authResp, err := s.generateTokens(userID, string(reg.UserType))
	if err != nil {
		return nil, err
	}

	// Update refresh token in database
	err = s.userRepo.UpdateRefreshToken(ctx, userID, authResp.RefreshToken)
	if err != nil {
		return nil, err
	}

	return authResp, nil
}

// RegisterDriver registers a new driver
func (s *AuthService) RegisterDriver(ctx context.Context, reg domain.Registration) (*domain.AuthResponse, error) {
	if reg.UserType != domain.UserTypeDriver {
		return nil, repository.ErrInvalidUserType
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		Email:     reg.Email,
		Password:  string(hashedPassword),
		UserType:  reg.UserType,
		Phone:     reg.Phone,
		FirstName: reg.FirstName,
		LastName:  reg.LastName,
		IsActive:  true,
	}

	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	authResp, err := s.generateTokens(userID, string(reg.UserType))
	if err != nil {
		return nil, err
	}

	err = s.userRepo.UpdateRefreshToken(ctx, userID, authResp.RefreshToken)
	if err != nil {
		return nil, err
	}

	return authResp, nil
}

func (s *AuthService) Login(ctx context.Context, login domain.Login) (*domain.AuthResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, login.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	authResp, err := s.generateTokens(user.ID, string(user.UserType))
	if err != nil {
		return nil, err
	}

	// Update refresh token in database
	err = s.userRepo.UpdateRefreshToken(ctx, user.ID, authResp.RefreshToken)
	if err != nil {
		return nil, err
	}

	return authResp, nil
}

func (s *AuthService) cacheAccessToken(ctx context.Context, userID, tokenID string, expiresIn time.Duration) error {
	if !s.useCache {
		return nil
	}

	key := fmt.Sprintf("access_token:%s:%s", userID, tokenID)
	return s.redisClient.Set(ctx, key, "valid", expiresIn).Err()
}

func (s *AuthService) getAccessToken(ctx context.Context, userID, tokenID string) (bool, error) {
	if !s.useCache {
		return false, nil
	}

	key := fmt.Sprintf("access_token:%s:%s", userID, tokenID)
	val, err := s.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return false, nil // Token not in cache
	} else if err != nil {
		return false, err // Redis error
	}

	return val == "valid", nil
}

func (s *AuthService) invalidateUserTokens(ctx context.Context, userID string) error {
	if !s.useCache {
		return nil
	}

	// Invalidate all user tokens by pattern
	pattern := fmt.Sprintf("access_token:%s:*", userID)
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return s.redisClient.Del(ctx, keys...).Err()
	}

	return nil
}

func (s *AuthService) blacklistToken(ctx context.Context, tokenID string, duration time.Duration) error {
	if !s.useCache {
		return nil
	}

	key := fmt.Sprintf("blacklist:%s", tokenID)
	return s.redisClient.Set(ctx, key, "revoked", duration).Err()
}

func (s *AuthService) isTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	if !s.useCache {
		return false, nil
	}

	key := fmt.Sprintf("blacklist:%s", tokenID)
	_, err := s.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return false, nil // Not blacklisted
	} else if err != nil {
		return false, err // Redis error
	}

	return true, nil // Token is blacklisted
}

// ValidateAccessToken validates an access token and returns the claims
func (s *AuthService) ValidateAccessToken(accessToken string) (jwt.MapClaims, error) {
	return s.validateToken(accessToken, s.jwtConfig.AccessTokenSecret)
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil
	}

	resetToken := uuid.New().String()

	fmt.Println(user.ID)
	return resetToken, nil
}

// generateTokens creates new access and refresh tokens
func (s *AuthService) generateTokens(userID, userType string) (*domain.AuthResponse, error) {
	// Create access token
	ctx := context.Background()
	tokenID := uuid.New().String()

	accessTokenExpiry := time.Now().Add(s.jwtConfig.AccessTokenExpiry)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"user_type": userType,
		"token_id":  tokenID,
		"exp":       accessTokenExpiry.Unix(),
		"iat":       time.Now().Unix(),
		"iss":       s.jwtConfig.Issuer,
	})

	accessTokenString, err := accessToken.SignedString([]byte(s.jwtConfig.AccessTokenSecret))
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshTokenExpiry := time.Now().Add(s.jwtConfig.RefreshTokenExpiry)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   userID,
		"user_type": userType,
		"token_id":  tokenID,
		"exp":       refreshTokenExpiry.Unix(),
		"iat":       time.Now().Unix(),
		"iss":       s.jwtConfig.Issuer,
	})

	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtConfig.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}

	if s.useCache {
		if err := s.cacheAccessToken(ctx, userID, tokenID, s.jwtConfig.AccessTokenExpiry); err != nil {
			fmt.Printf("Warning: Failed to cache token: %v\n", err)
		}
	}

	return &domain.AuthResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
		UserID:       userID,
		UserType:     userType,
	}, nil
}

func (s *AuthService) validateToken(tokenString, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	// Check if token is valid
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check if token is blacklisted
	if s.useCache {
		tokenID, ok := claims["token_id"].(string)
		if !ok {
			return nil, ErrInvalidToken
		}

		blacklisted, err := s.isTokenBlacklisted(context.Background(), tokenID)
		if err != nil {
			fmt.Printf("Warning: Failed to check token blacklist: %v\n", err)
		} else if blacklisted {
			return nil, ErrInvalidToken
		}
	}

	return claims, nil
}

// Update refresh token to invalidate old tokens
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
	// Parse and validate refresh token
	claims, err := s.validateToken(refreshToken, s.jwtConfig.RefreshTokenSecret)
	if err != nil {
		return nil, err
	}

	// Extract user ID and type from claims
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userType, ok := claims["user_type"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	tokenID, ok := claims["token_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Get user from database
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Verify that the stored refresh token matches
	if user.RefreshToken != refreshToken {
		return nil, ErrInvalidToken
	}

	// Blacklist the current token
	if s.useCache && tokenID != "" {
		if err := s.blacklistToken(ctx, tokenID, s.jwtConfig.RefreshTokenExpiry); err != nil {
			fmt.Printf("Warning: Failed to blacklist token: %v\n", err)
		}
	}

	// Generate new tokens
	authResp, err := s.generateTokens(userID, userType)
	if err != nil {
		return nil, err
	}

	// Update refresh token in database
	err = s.userRepo.UpdateRefreshToken(ctx, userID, authResp.RefreshToken)
	if err != nil {
		return nil, err
	}

	return authResp, nil
}

// Update ResetPassword to invalidate all tokens
func (s *AuthService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	// Get user by ID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	user.RefreshToken = "" // Invalidate all existing sessions

	// Invalidate all tokens in cache
	if s.useCache {
		if err := s.invalidateUserTokens(ctx, userID); err != nil {
			fmt.Printf("Warning: Failed to invalidate user tokens in cache: %v\n", err)
		}
	}

	// Save updated user
	return s.userRepo.UpdateUser(ctx, *user)
}
