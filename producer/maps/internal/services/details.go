package services

import (
	"context"
	"log"
	"time"

	"googlemaps.github.io/maps"
	"navik/producer/maps/internal/models"
)

// PlaceDetailsService handles place details operations
type PlaceDetailsService struct {
	client   *maps.Client
	infoLog  *log.Logger
	errorLog *log.Logger
	debugLog *log.Logger
}

// NewPlaceDetailsService creates a new place details service
func NewPlaceDetailsService(client *maps.Client, infoLog, errorLog, debugLog *log.Logger) *PlaceDetailsService {
	return &PlaceDetailsService{
		client:   client,
		infoLog:  infoLog,
		errorLog: errorLog,
		debugLog: debugLog,
	}
}

// GetPlaceDetails retrieves detailed information about a place
func (s *PlaceDetailsService) GetPlaceDetails(placeID string) (*models.PlaceDetails, error) {
	s.infoLog.Printf("Fetching place details for ID: %s", placeID)

	// Create place details request
	r := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskName,
			maps.PlaceDetailsFieldMaskFormattedAddress,
			maps.PlaceDetailsFieldMaskGeometry,
		},
	}

	// Call the Places API for details
	startTime := time.Now()
	resp, err := s.client.PlaceDetails(context.Background(), r)
	if err != nil {
		s.errorLog.Printf("Google Maps Place Details API error: %v", err)
		return nil, err
	}

	s.debugLog.Printf("Google Maps Place Details API response time: %v", time.Since(startTime))

	// Convert API response to our PlaceDetails format
	details := &models.PlaceDetails{
		ID:      resp.PlaceID,
		Name:    resp.Name,
		Address: resp.FormattedAddress,
	}

	// Add location if available
	if resp.Geometry.Location.Lat != 0 || resp.Geometry.Location.Lng != 0 {
		details.Latitude = resp.Geometry.Location.Lat
		details.Longitude = resp.Geometry.Location.Lng
	}

	return details, nil
}
