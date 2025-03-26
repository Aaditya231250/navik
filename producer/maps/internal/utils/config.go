package utils

import (
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
	"os"
)

// Config holds application configuration
type Config struct {
	MapsClient *maps.Client
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	// Get API key from environment
	apiKey := os.Getenv("GMAPS_API_KEY")
	if apiKey == "" {
		return nil, err
	}

	// Initialize Google Maps client
	mapsClient, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &Config{
		MapsClient: mapsClient,
	}, nil
}
