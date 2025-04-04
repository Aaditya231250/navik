package service

import (
    "context"
    "errors"
    "fmt"
    "time"
    
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

type JWTConfig struct {
    AccessTokenSecret  string
    RefreshTokenSecret string
    AccessTokenExpiry  time.Duration
    RefreshTokenExpiry time.Duration
    Issuer             string
}

type AuthService struct {
    userRepo *repository.UserRepository
    jwtConfig JWTConfig
}

func NewAuthService(userRepo *repository.UserRepository, jwtConfig JWTConfig) *AuthService {
    return &AuthService{
        userRepo:  userRepo,
        jwtConfig: jwtConfig,
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
    
    // Create user with driver-specific fields
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

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, login domain.Login) (*domain.AuthResponse, error) {
    // Get user by email
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

// RefreshToken generates new access token using refresh token
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
    
    // Get user from database
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // Verify that the stored refresh token matches
    if user.RefreshToken != refreshToken {
        return nil, ErrInvalidToken
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
    
    // Save updated user
    return s.userRepo.UpdateUser(ctx, *user)
}

// generateTokens creates new access and refresh tokens
func (s *AuthService) generateTokens(userID, userType string) (*domain.AuthResponse, error) {
    // Create access token
    accessTokenExpiry := time.Now().Add(s.jwtConfig.AccessTokenExpiry)
    accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id":   userID,
        "user_type": userType,
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
        "exp":       refreshTokenExpiry.Unix(),
        "iat":       time.Now().Unix(),
        "iss":       s.jwtConfig.Issuer,
    })
    
    refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtConfig.RefreshTokenSecret))
    if err != nil {
        return nil, err
    }
    
    return &domain.AuthResponse{
        AccessToken:  accessTokenString,
        RefreshToken: refreshTokenString,
        ExpiresIn:    int(s.jwtConfig.AccessTokenExpiry.Seconds()),
        UserID:       userID,
        UserType:     userType,
    }, nil
}

// validateToken parses and validates a JWT token
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
    
    return claims, nil
}
