package repository

import (
	"context"
	"sync"
	"fmt"

	"location-service/internal/model"
)


type LocationRepository interface {
	Store(ctx context.Context, loc model.Location) error
	GetLatest(ctx context.Context, driverID string) (model.Location, error)
	GetByCity(ctx context.Context, city string, limit int) ([]model.Location, error)
}

// InMemoryLocationRepository implements the LocationRepository interface with in-memory storage
type InMemoryLocationRepository struct {
	locations map[string]model.Location
	cityIndex map[string][]string // city -> list of driver IDs
	mutex     sync.RWMutex
}

// NewInMemoryLocationRepository creates a new in-memory repository
func NewInMemoryLocationRepository() *InMemoryLocationRepository {
	return &InMemoryLocationRepository{
		locations: make(map[string]model.Location),
		cityIndex: make(map[string][]string),
	}
}

// Store saves a location update in memory
func (r *InMemoryLocationRepository) Store(ctx context.Context, loc model.Location) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	// Update location
	r.locations[loc.DriverID] = loc
	
	// Update city index
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

// GetLatest retrieves the latest location for a driver
func (r *InMemoryLocationRepository) GetLatest(ctx context.Context, driverID string) (model.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	loc, exists := r.locations[driverID]
	if !exists {
		return model.Location{}, fmt.Errorf("driver location not found")
	}
	
	return loc, nil
}

// GetByCity retrieves locations for drivers in a specific city
func (r *InMemoryLocationRepository) GetByCity(ctx context.Context, city string, limit int) ([]model.Location, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	driverIDs, exists := r.cityIndex[city]
	if !exists {
		return []model.Location{}, nil
	}
	
	var locations []model.Location
	for _, id := range driverIDs {
		if loc, ok := r.locations[id]; ok {
			locations = append(locations, loc)
			if len(locations) >= limit {
				break
			}
		}
	}
	
	return locations, nil
}
