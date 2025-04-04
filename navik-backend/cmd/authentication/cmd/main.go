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
	// Load configuration
	cfg := config.LoadConfig()

	// Configure AWS with static credentials
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

	// Configure DynamoDB client for local development
	var dynamoClient *dynamodb.Client
	if cfg.DynamoDBEndpoint != "" {
		// Use custom endpoint for local development
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

	// Create DynamoDB table if it doesn't exist
	err = repository.CreateUsersTable(context.Background(), dynamoClient, cfg.DynamoDBTableName)
	if err != nil {
		log.Fatalf("failed to create users table: %v", err)
	}

	// Initialize repository
	userRepo := repository.NewUserRepository(dynamoClient, cfg.DynamoDBTableName)

	// Initialize services
	jwtConfig := service.JWTConfig{
		AccessTokenSecret:  cfg.JWTAccessSecret,
		RefreshTokenSecret: cfg.JWTRefreshSecret,
		AccessTokenExpiry:  cfg.JWTAccessExpiry,
		RefreshTokenExpiry: cfg.JWTRefreshExpiry,
		Issuer:             cfg.JWTIssuer,
	}
	authService := service.NewAuthService(userRepo, jwtConfig)

	// Add the profile service
	profileService := service.NewProfileService(userRepo)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Add the profile handler
	profileHandler := handlers.NewProfileHandler(profileService)

	// Create router
	router := mux.NewRouter()

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/refresh", authHandler.RefreshToken).Methods("POST")
	router.HandleFunc("/api/auth/password-reset/request", authHandler.RequestPasswordReset).Methods("POST")

	// Protected routes
	api := router.PathPrefix("/api").Subrouter()
	api.Use(handlers.AuthMiddleware(authService))

	// Customer-only routes
	customerAPI := api.PathPrefix("/customer").Subrouter()
	customerAPI.Use(handlers.RoleMiddleware(string(domain.UserTypeCustomer)))

	// Add customer profile endpoints
	customerAPI.HandleFunc("/profile", profileHandler.GetCustomerProfile).Methods("GET")
	customerAPI.HandleFunc("/profile", profileHandler.UpdateCustomerProfile).Methods("PUT")

	// Driver-only routes
	driverAPI := api.PathPrefix("/driver").Subrouter()
	driverAPI.Use(handlers.RoleMiddleware(string(domain.UserTypeDriver)))

	// Add driver profile endpoints
	driverAPI.HandleFunc("/profile", profileHandler.GetDriverProfile).Methods("GET")
	driverAPI.HandleFunc("/profile", profileHandler.UpdateDriverProfile).Methods("PUT")

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
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
