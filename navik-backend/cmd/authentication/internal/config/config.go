package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server
	ServerPort string

	// DynamoDB
	DynamoDBTableName  string
	DynamoDBEndpoint   string
	AWSRegion          string
	AwsAccessKeyID     string
	AwsSecretAccessKey string

	// JWT
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration
	JWTIssuer        string
}

func LoadConfig() *Config {
	accessExpiry, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	refreshExpiry, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))

	return &Config{
		// Server
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// DynamoDB
		DynamoDBTableName:  getEnv("DYNAMODB_TABLE_NAME", "Users"),
		DynamoDBEndpoint:   getEnv("DYNAMODB_ENDPOINT", "http://dynamodb-local:8000"),
		AWSRegion:          getEnv("AWS_REGION", "local"),
		AwsAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", "localkey"),
		AwsSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", "localsecret"),

		// JWT
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
		JWTAccessExpiry:  time.Duration(accessExpiry) * time.Minute,
		JWTRefreshExpiry: time.Duration(refreshExpiry) * 24 * time.Hour,
		JWTIssuer:        getEnv("JWT_ISSUER", "auth-service"),
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
