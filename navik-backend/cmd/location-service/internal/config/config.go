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

	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		config.Kafka.Brokers = strings.Split(brokers, ",")
	}

	return &config, nil
}
