package services

import (
	"context"
	"log"
	"time"

	"internal-maps/internal/models"

	"googlemaps.github.io/maps"
)

// GeocodingService handles geocoding operations
type GeocodingService struct {
	client   *maps.Client
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// NewGeocodingService creates a new geocoding service
func NewGeocodingService(client *maps.Client, infoLog, errorLog, debugLog *log.Logger) *GeocodingService {
	return &GeocodingService{
		client:   client,
		infoLog:  infoLog,
		errorLog: errorLog,
		debugLog: debugLog,
	}
}

// SearchPlaces searches for places based on a query string
func (s *GeocodingService) SearchPlaces(query string) ([]models.Place, error) {
	s.debugLog.Printf("Received search query: %s", query)
	var results []models.Place

	if query == "" {
		return results, nil
	}

	// Create autocomplete request
	r := &maps.PlaceAutocompleteRequest{
		Input: query,
	}

	// Call the Places API
	startTime := time.Now()
	s.infoLog.Printf("Sending request to Google Maps API for query: %s", query)
	resp, err := s.client.PlaceAutocomplete(context.Background(), r)
	if err != nil {
		s.errorLog.Printf("Google Maps API error: %v", err)
		return nil, err
	}

	s.debugLog.Printf("Google Maps API response time: %v", time.Since(startTime))
	s.debugLog.Printf("Received %d predictions from Google Maps API", len(resp.Predictions))

	// Convert Google Maps predictions to our Place format
	for _, prediction := range resp.Predictions {
		place := models.Place{
			ID:      prediction.PlaceID,
			Name:    prediction.StructuredFormatting.MainText,
			Address: prediction.Description,
		}
		results = append(results, place)
	}

	return results, nil
}
