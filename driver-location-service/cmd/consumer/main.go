package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"driver-location-service/internal/model"
	"driver-location-service/pkg/kafka"
)

type Config struct {
	Kafka struct {
		Brokers []string `json:"brokers"`
		GroupID string   `json:"group_id"`
		Topics  []string `json:"topics"`
	} `json:"kafka"`
}

func main() {
	// Load configuration
	var config Config
	if err := loadConfig("config.json", &config); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create consumer with a handler function
	consumer, err := kafka.NewConsumer(
		config.Kafka.Brokers,
		config.Kafka.GroupID,
		config.Kafka.Topics,
		handleLocation,
	)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Start consuming messages
	log.Println("Starting consumer...")
	if err := consumer.Consume(context.Background()); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
}

// loadConfig loads the application configuration from a JSON file
func loadConfig(filename string, config *Config) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(file, config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	return nil
}


func handleLocation(loc model.Location) error {
	log.Printf("Processing location: Driver %s in %s at (%f, %f)",
		loc.DriverID, loc.City, loc.Latitude, loc.Longitude)


	return nil
}
