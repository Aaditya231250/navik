package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Location struct {
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

const (
	logDir        = "./logs"
	consumerGroup = "location-file-writer"
)

func main() {
	// Create log directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	// Kafka consumer configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers":  "localhost:9101,localhost:9102,localhost:9103",
		"group.id":           consumerGroup,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}

	// Create consumer
	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create consumer: %v", err))
	}
	defer consumer.Close()

	// Subscribe to all regional topics
	err = consumer.SubscribeTopics([]string{"mumbai-locations", "pune-locations", "delhi-locations"}, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to subscribe to topics: %v", err))
	}

	// Handle shutdown
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	for {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			return
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				continue
			}

			// Process message
			if err := processMessage(msg); err == nil {
				// Commit offset only if processing succeeds
				consumer.CommitMessage(msg)
			}
		}
	}
}

func processMessage(msg *kafka.Message) error {
	var loc Location
	if err := json.Unmarshal(msg.Value, &loc); err != nil {
		return fmt.Errorf("failed to unmarshal location: %w", err)
	}

	// Determine city from topic name
	city := getCityFromTopic(*msg.TopicPartition.Topic)
	if city == "" {
		return fmt.Errorf("unknown topic: %s", *msg.TopicPartition.Topic)
	}

	// Write to city-specific log file
	return writeToFile(city, loc)
}

func getCityFromTopic(topic string) string {
	switch topic {
	case "mumbai-locations":
		return "mumbai"
	case "pune-locations":
		return "pune"
	case "delhi-locations":
		return "delhi"
	default:
		return ""
	}
}

func writeToFile(city string, loc Location) error {
	fpath := filepath.Join(logDir, fmt.Sprintf("%s_locations.log", city))
	file, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	entry := fmt.Sprintf("[%s] %s: (%.6f, %.6f)\n",
		time.Unix(loc.Timestamp, 0).Format(time.RFC3339),
		loc.City,
		loc.Latitude,
		loc.Longitude)

	if _, err := file.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}
