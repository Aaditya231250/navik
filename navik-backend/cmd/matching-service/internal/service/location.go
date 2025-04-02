package service

import (
	"context"
	"fmt"
	"log"

	"matching-service/internal/model"
	"matching-service/pkg/kafka"
)

type LocationService interface {
	UpdateLocation(ctx context.Context, loc model.UserLocation) error
}

type locationService struct {
	producer *kafka.Producer
}

func NewLocationService(repository interface{}, producer *kafka.Producer) LocationService {
	return &locationService{
		producer: producer,
	}
}

func (s *locationService) UpdateLocation(ctx context.Context, loc model.UserLocation) error {
	if err := loc.Validate(); err != nil {
		return fmt.Errorf("invalid location data: %w", err)
	}

	if s.producer != nil {
		if err := s.producer.SendLocation(loc); err != nil {
			log.Printf("Warning: Failed to publish location to Kafka: %v", err)
			return fmt.Errorf("failed to publish location: %w", err)
		}
	}

	return nil
}
