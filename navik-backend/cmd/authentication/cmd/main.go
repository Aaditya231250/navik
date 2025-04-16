package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gorilla/mux"

	"authentication/internal/config"
	"authentication/internal/domain"
	"authentication/internal/handlers"
	"authentication/internal/repository"
	"authentication/internal/service"
)

func main() {
	cfg := config.LoadConfig()

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AwsAccessKeyID,
			cfg.AwsSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	var dynamoClient *dynamodb.Client
	if cfg.DynamoDBEndpoint != "" {
		dynamoClient = dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
			o.EndpointResolver = dynamodb.EndpointResolverFunc(
				func(region string, options dynamodb.EndpointResolverOptions) (aws.Endpoint, error) {
					return aws.Endpoint{
						URL:               cfg.DynamoDBEndpoint,
						HostnameImmutable: true,
						SigningRegion:     cfg.AWSRegion,
					}, nil
				},
			)
		})
	} else {
		dynamoClient = dynamodb.NewFromConfig(awsCfg)
	}

	err = repository.CreateUsersTable(context.Background(), dynamoClient, cfg.DynamoDBTableName)
	if err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}

	userRepo := repository.NewUserRepository(dynamoClient, cfg.DynamoDBTableName)

	redisConfig := service.RedisConfig{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
		Enabled:  true,
	}

	// Initialize services
	jwtConfig := service.JWTConfig{
		AccessTokenSecret:   cfg.JWTAccessSecret,
		RefreshTokenSecret:  cfg.JWTRefreshSecret,
		AccessTokenExpiry:   cfg.JWTAccessExpiry,
		RefreshTokenExpiry:  cfg.JWTRefreshExpiry,
		Issuer:              cfg.JWTIssuer,
		AccessTokenCacheTTL: 60 * 2,
	}
	authService := service.NewAuthService(userRepo, jwtConfig, redisConfig)

	profileService := service.NewProfileService(userRepo)

	authHandler := handlers.NewAuthHandler(authService)

	profileHandler := handlers.NewProfileHandler(profileService)

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST")
	router.HandleFunc("/api/auth/password-reset/request", authHandler.RequestPasswordReset).Methods("POST")

	api := router.PathPrefix("/api").Subrouter()
	api.Use(handlers.AuthMiddleware(authService))

	customerAPI := api.PathPrefix("/customer").Subrouter()
	customerAPI.Use(handlers.RoleMiddleware(string(domain.UserTypeCustomer)))

	customerAPI.HandleFunc("/profile", profileHandler.GetCustomerProfile).Methods("GET")
	customerAPI.HandleFunc("/profile", profileHandler.UpdateCustomerProfile).Methods("PUT")

	driverAPI := api.PathPrefix("/driver").Subrouter()
	driverAPI.Use(handlers.RoleMiddleware(string(domain.UserTypeDriver)))

	driverAPI.HandleFunc("/profile", profileHandler.GetDriverProfile).Methods("GET")
	driverAPI.HandleFunc("/profile", profileHandler.UpdateDriverProfile).Methods("PUT")

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}
