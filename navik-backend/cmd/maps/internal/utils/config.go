package utils

import (
	"errors"
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
	"os"
)

type Config struct {
	MapsClient *maps.Client
}

func LoadConfig() (*Config, error) {

	_ = godotenv.Load() 

	apiKey := os.Getenv("GMAPS_API_KEY")
	if apiKey == "" {
		return nil, errors.New("GMAPS_API_KEY environment variable is not set")
	}

	mapsClient, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	return &Config{
		MapsClient: mapsClient,
	}, nil
}
