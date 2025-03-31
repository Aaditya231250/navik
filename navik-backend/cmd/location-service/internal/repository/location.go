package repository

import (
	"context"
	"sync"

	"location-service/internal/model"
)

type LocationRepository interface {
	Store(ctx context.Context, loc model.Location) error
}

type InMemoryLocationRepository struct {
	locations map[string]model.Location
	cityIndex map[string][]string // city -> list of driver IDs
	mutex     sync.RWMutex
}

func NewInMemoryLocationRepository() *InMemoryLocationRepository {
	return &InMemoryLocationRepository{
		locations: make(map[string]model.Location),
		cityIndex: make(map[string][]string),
	}
}

func (r *InMemoryLocationRepository) Store(ctx context.Context, loc model.Location) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.locations[loc.DriverID] = loc

	found := false
	drivers, exists := r.cityIndex[loc.City]
	if !exists {
		r.cityIndex[loc.City] = []string{loc.DriverID}
	} else {
		for _, id := range drivers {
			if id == loc.DriverID {
				found = true
				break
			}
		}
		if !found {
			r.cityIndex[loc.City] = append(r.cityIndex[loc.City], loc.DriverID)
		}
	}

	return nil
}
