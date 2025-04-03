package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"driver-location-service/internal/model"
	"driver-location-service/pkg/kafka"
)

// Config holds the application configuration
type Config struct {
	Kafka struct {
		Brokers     []string `json:"brokers"`
		TopicFormat string   `json:"topic_format"`
	} `json:"kafka"`
	Server struct {
		Port int `json:"port"`
	} `json:"server"`
}

var (
	producer *kafka.Producer
	config   Config
)

func main() {
	// Load configuration
	if err := loadConfig("config.json"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Kafka producer
	var err error
	producer, err = kafka.NewProducer(config.Kafka.Brokers, config.Kafka.TopicFormat)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Set up HTTP routes
	http.HandleFunc("/api/location", handleLocationUpdate)

	// Start the server
	serverAddr := fmt.Sprintf(":%d", config.Server.Port)
	log.Printf("Starting server on %s", serverAddr)
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	log.Println("Server stopped")
}

// loadConfig loads the application configuration from a JSON file
func loadConfig(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}

	if err := json.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("error parsing config file: %w", err)
	}

	return nil
}

func handleLocationUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loc model.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if loc.DriverID == "" || loc.City == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Set timestamp if not provided
	if loc.Timestamp == 0 {
		loc.Timestamp = time.Now().Unix()
	}

	// Send to Kafka
	log.Printf("Received location update: %+v", loc)
	if err := producer.SendToProducer(loc); err != nil {
		log.Printf("Error sending to Kafka: %v", err)
		http.Error(w, "Failed to process location update", http.StatusInternalServerError)
		return
	}
	log.Printf("Location update sent to Kafka: %+v", loc)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Location update processed",
	})
}
