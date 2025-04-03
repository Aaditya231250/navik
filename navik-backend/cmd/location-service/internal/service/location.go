package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"location-service/internal/model"
	"location-service/internal/repository"
	"location-service/pkg/kafka"
)

type LocationService interface {
	UpdateLocation(ctx context.Context, loc model.Location) error
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
	if err := loc.Validate(); err != nil {
		return fmt.Errorf("invalid location data: %w", err)
	}

	if s.producer != nil {
		if err := s.producer.SendToProducer(loc, loc.City, ""); err != nil {
			log.Printf("Warning: Failed to publish location to Kafka: %v", err)
		}
	}

	return nil
}

func (s *locationService) ProcessLocationUpdate(loc model.Location) error {
	ctx := context.Background()

	if time.Now().Unix()-loc.Timestamp > 300 { // Older than 5 minutes
		log.Printf("Skipping stale location update for driver %s (%.2f minutes old)",
			loc.DriverID, float64(time.Now().Unix()-loc.Timestamp)/60)
		return nil
	}

	log.Printf("Processing location update for driver %s in %s",
		loc.DriverID, loc.City)

	return s.repository.Store(ctx, loc)
}
