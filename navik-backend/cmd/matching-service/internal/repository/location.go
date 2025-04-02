package repository

import (
	"context"
	"sync"

	"matching-service/internal/model"
)

type UserLocationRepository interface {
	Store(ctx context.Context, loc model.UserLocation) error
}

type InMemoryUserLocationRepository struct {
	UserLocations map[string]model.UserLocation
	cityIndex     map[string][]string // city -> list of driver IDs
	mutex         sync.RWMutex
}

func NewInMemoryUserLocationRepository() *InMemoryUserLocationRepository {
	return &InMemoryUserLocationRepository{
		UserLocations: make(map[string]model.UserLocation),
		cityIndex:     make(map[string][]string),
	}
}

func (r *InMemoryUserLocationRepository) Store(ctx context.Context, loc model.UserLocation) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.UserLocations[loc.UserID] = loc

	found := false
	drivers, exists := r.cityIndex[loc.City]
	if !exists {
		r.cityIndex[loc.City] = []string{loc.UserID}
	} else {
		for _, id := range drivers {
			if id == loc.UserID {
				found = true
				break
			}
		}
		if !found {
			r.cityIndex[loc.City] = append(r.cityIndex[loc.City], loc.UserID)
		}
	}

	return nil
}
