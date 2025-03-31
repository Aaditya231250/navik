// internal/service/location.go
package service

import (
	"context"
	"fmt"
	"log"

	"location-service/internal/model"
	"location-service/internal/repository"
	"location-service/pkg/kafka"
)

type LocationService interface {
	UpdateLocation(ctx context.Context, loc model.Location) error
	GetDriverLocation(ctx context.Context, driverID string) (model.Location, error)
	GetDriversByCity(ctx context.Context, city string, limit int) ([]model.Location, error)
	ProcessLocationUpdate(loc model.Location) error
}

type locationService struct {
	repository repository.LocationRepository
	producer   *kafka.Producer
}

func NewLocationService(repo repository.LocationRepository, producer *kafka.Producer) LocationService {
	return &locationService{
		repository: repo,
		producer:   producer,
	}
}

func (s *locationService) UpdateLocation(ctx context.Context, loc model.Location) error {
	// Validate location data
	if err := loc.Validate(); err != nil {
		return fmt.Errorf("invalid location data: %w", err)
	}

	// Store in the in-memory repository
	if err := s.repository.Store(ctx, loc); err != nil {
		return fmt.Errorf("failed to store location: %w", err)
	}

	if s.producer != nil {
		if err := s.producer.SendLocation(loc); err != nil {
			log.Printf("Warning: Failed to publish location to Kafka: %v", err)
			// Don't fail the request if Kafka publish fails
		}
	}

	return nil
}

func (s *locationService) GetDriverLocation(ctx context.Context, driverID string) (model.Location, error) {
	return s.repository.GetLatest(ctx, driverID)
}

func (s *locationService) GetDriversByCity(ctx context.Context, city string, limit int) ([]model.Location, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100 // Default limit
	}
	return s.repository.GetByCity(ctx, city, limit)
}

func (s *locationService) ProcessLocationUpdate(loc model.Location) error {
	return s.repository.Store(context.Background(), loc)
}
