package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Kafka struct {
		Brokers     []string `json:"brokers"`
		TopicFormat string   `json:"topic_format"`
		GroupID     string   `json:"group_id"`
		Topics      []string `json:"topics"`
	} `json:"kafka"`
	Server struct {
		Port int `json:"port"`
	} `json:"server"`
	DynamoDB struct {
		Endpoint  string `json:"endpoint"`
		Region    string `json:"region"`
		TableName string `json:"table_name"`
		AccessKey string `json:"access_key"`
		SecretKey string `json:"secret_key"`
	} `json:"dynamodb"`
	Matching struct {
		MinDriversToReturn int     `json:"min_drivers_to_return"`
		MaxDistanceKm      float64 `json:"max_distance_km"`
	} `json:"matching"`
}

// Load loads configuration from environment variables or a file
func Load(filename string) (*Config, error) {
	var config Config

	// Load from file
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Override with environment variables if provided
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.Kafka.Brokers = strings.Split(brokers, ",")
	}

	// Set defaults
	if len(config.Kafka.Brokers) == 0 {
		config.Kafka.Brokers = []string{"localhost:9092"}
	}

	if config.Kafka.TopicFormat == "" {
		config.Kafka.TopicFormat = "%s-users"
	}

	if config.Kafka.GroupID == "" {
		config.Kafka.GroupID = "matching-service"
	}

	if len(config.Kafka.Topics) == 0 {
		config.Kafka.Topics = []string{"mumbai-users", "pune-users", "delhi-users"}
	}

	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	if config.DynamoDB.Endpoint == "" {
		config.DynamoDB.Endpoint = "http://localhost:8000"
	}

	if config.DynamoDB.Region == "" {
		config.DynamoDB.Region = "us-west-2"
	}

	if config.DynamoDB.TableName == "" {
		config.DynamoDB.TableName = "DriverLocations"
	}

	if config.Matching.MinDriversToReturn == 0 {
		config.Matching.MinDriversToReturn = 5
	}

	if config.Matching.MaxDistanceKm == 0 {
		config.Matching.MaxDistanceKm = 10.0
	}

	return &config, nil
}
