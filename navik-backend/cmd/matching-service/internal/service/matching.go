package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"time"
	"encoding/json"

	"matching-service/internal/model"
	"matching-service/internal/repository"
	"matching-service/internal/util"
	"github.com/go-redis/redis/v8"
)

type MatchingService interface {
	ProcessUserLocation(ctx context.Context, loc model.UserLocation) error
}

type matchingService struct {
	repository         repository.DriverRepository
	minDriversToReturn int
	maxDistanceKm      float64
	redisClient        *redis.Client
}
// NewMatchingService creates a new matching service
func NewMatchingService(repo repository.DriverRepository, config struct {
	MinDriversToReturn int
	MaxDistanceKm      float64
}, redisClient *redis.Client) MatchingService {
	return &matchingService{
		repository:         repo,
		minDriversToReturn: config.MinDriversToReturn,
		maxDistanceKm:      config.MaxDistanceKm,
		redisClient:        redisClient,
	}
}

// ProcessUserLocation processes a user location message
func (s *matchingService) ProcessUserLocation(ctx context.Context, loc model.UserLocation) error {
	// Enrich with H3 indices
	enrichedUser := s.enrichUserLocation(loc)

	log.Printf("Received user request: %s at H3-9: %s",
		enrichedUser.UserID, enrichedUser.H3Index9)

	// Trigger the matching algorithm
	drivers, err := s.findDriversForUser(ctx, enrichedUser)
	if err != nil {
		return fmt.Errorf("error finding drivers: %w", err)
	}

	// Process the matching results
	s.processMatchingResults(enrichedUser, drivers)
	// Store the user location in Redis for future reference
	response := s.formatDriverResponse(enrichedUser, drivers)
	data, _ := json.Marshal(response)
	// Publish to Redis Pub/Sub for this user
	if err := s.redisClient.Publish(ctx, "user:"+loc.UserID, data).Err(); err != nil {
		log.Printf("Failed to publish to Redis: %v", err)
	}
	return nil

}

// enrichUserLocation adds H3 indices to a user location
func (s *matchingService) enrichUserLocation(loc model.UserLocation) model.EnrichedUserLocation {
	h3Index9 := util.GeoToH3Index(loc.Latitude, loc.Longitude, 9)
	h3Index8 := util.GeoToH3Index(loc.Latitude, loc.Longitude, 8)
	h3Index7 := util.GeoToH3Index(loc.Latitude, loc.Longitude, 7)

	return model.EnrichedUserLocation{
		UserLocation: loc,
		H3Index9:     h3Index9,
		H3Index8:     h3Index8,
		H3Index7:     h3Index7,
	}
}


func (s *matchingService) findDriversForUser(ctx context.Context, user model.EnrichedUserLocation) ([]model.DriverLocation, error) {
	// Step 1: Try to find drivers in the exact H9 cell
	drivers, err := s.repository.FindDriversInH9Cell(ctx, user.H3Index9)
	if err != nil {
		return nil, fmt.Errorf("error querying H9 cell: %w", err)
	}

	log.Printf("Found %d drivers in exact H9 cell %s", len(drivers), user.H3Index9)

	// If we found enough drivers, rank and return them
	if len(drivers) >= s.minDriversToReturn {
		rankedDrivers := s.assignRandomDistances(drivers)
		return s.getTopDrivers(rankedDrivers, s.minDriversToReturn), nil
	}

	// Step 2: Not enough drivers, try H9 neighbors
	h9Neighbors := util.GetH3Neighbors(user.H3Index9)
	log.Printf("Looking for drivers in %d neighboring H9 cells", len(h9Neighbors))

	// Query for drivers in neighboring H9 cells
	neighborDrivers, err := s.repository.FindDriversInH9Cells(ctx, h9Neighbors)
	if err != nil {
		return nil, fmt.Errorf("error querying H9 neighbor cells: %w", err)
	}

	// Combine with drivers from the exact cell
	allDrivers := append(drivers, neighborDrivers...)
	log.Printf("Found total of %d drivers in H9 cell and neighbors", len(allDrivers))

	// If we found enough drivers, rank and return them
	if len(allDrivers) >= s.minDriversToReturn {
		rankedDrivers := s.assignRandomDistances(allDrivers)
		return s.getTopDrivers(rankedDrivers, s.minDriversToReturn), nil
	}

	// Step 3: Still not enough drivers, move up to H8 cell
	log.Printf("Not enough drivers in H9 cells, moving up to H8 cell %s", user.H3Index8)

	h8Drivers, err := s.repository.FindDriversInH8Cell(ctx, user.H3Index8)
	if err != nil {
		return nil, fmt.Errorf("error querying H8 cell: %w", err)
	}

	log.Printf("Found %d drivers in H8 cell", len(h8Drivers))

	// Filter out drivers we already found in H9 cells to avoid duplicates
	h8Drivers = s.filterOutDuplicateDrivers(h8Drivers, allDrivers)

	allDrivers = append(allDrivers, h8Drivers...)

	if len(allDrivers) >= s.minDriversToReturn {
		rankedDrivers := s.assignRandomDistances(allDrivers)
		return s.getTopDrivers(rankedDrivers, s.minDriversToReturn), nil
	}

	// Step 4: Still not enough drivers, try H8 neighbors
	h8Neighbors := util.GetH3Neighbors(user.H3Index8)
	log.Printf("Looking for drivers in %d neighboring H8 cells", len(h8Neighbors))

	// Query for drivers in neighboring H8 cells
	h8NeighborDrivers, err := s.repository.FindDriversInH8Cells(ctx, h8Neighbors)
	if err != nil {
		return nil, fmt.Errorf("error querying H8 neighbor cells: %w", err)
	}

	// Filter out drivers we already found
	h8NeighborDrivers = s.filterOutDuplicateDrivers(h8NeighborDrivers, allDrivers)

	// Combine with all previous drivers
	allDrivers = append(allDrivers, h8NeighborDrivers...)
	log.Printf("Found total of %d drivers after H8 neighbors", len(allDrivers))

	// If we found enough drivers, rank and return them
	if len(allDrivers) >= s.minDriversToReturn {
		rankedDrivers := s.assignRandomDistances(allDrivers)
		return s.getTopDrivers(rankedDrivers, s.minDriversToReturn), nil
	}

	// Step 5: Still not enough drivers, move up to H7 cell
	log.Printf("Not enough drivers in H8 cells, moving up to H7 cell %s", user.H3Index7)

	h7Drivers, err := s.repository.FindDriversInH7Cell(ctx, user.H3Index7)
	if err != nil {
		return nil, fmt.Errorf("error querying H7 cell: %w", err)
	}

	log.Printf("Found %d drivers in H7 cell", len(h7Drivers))

	h7Drivers = s.filterOutDuplicateDrivers(h7Drivers, allDrivers)

	allDrivers = append(allDrivers, h7Drivers...)

	if len(allDrivers) >= s.minDriversToReturn {
		rankedDrivers := s.assignRandomDistances(allDrivers)
		return s.getTopDrivers(rankedDrivers, s.minDriversToReturn), nil
	}

	// Return whatever drivers we found, even if less than minDriversToReturn
	if len(allDrivers) > 0 {
		rankedDrivers := s.assignRandomDistances(allDrivers)
		return rankedDrivers, nil
	}

	// No drivers found
	return []model.DriverLocation{}, nil
}

// assignRandomDistances assigns random distances to drivers and sorts them
func (s *matchingService) assignRandomDistances(drivers []model.DriverLocation) []model.DriverLocation {
	// Assign random distances
	for i := range drivers {
		// Random distance between 0.5 and 5 km
		drivers[i].Distance = 0.5 + rand.Float64()*4.5

		// Calculate estimated time of arrival (1 minute per km + random 1-3 minutes)
		drivers[i].ETA = int(drivers[i].Distance) + rand.Intn(3) + 1
	}

	// Sort by distance
	sort.Slice(drivers, func(i, j int) bool {
		return drivers[i].Distance < drivers[j].Distance
	})

	return drivers
}

// getTopDrivers returns the top N drivers from a ranked list
func (s *matchingService) getTopDrivers(drivers []model.DriverLocation, n int) []model.DriverLocation {
	if len(drivers) <= n {
		return drivers
	}
	return drivers[:n]
}

// filterOutDuplicateDrivers removes drivers that are already in the existingDrivers list
func (s *matchingService) filterOutDuplicateDrivers(newDrivers, existingDrivers []model.DriverLocation) []model.DriverLocation {
	// Create a map of existing driver IDs for quick lookup
	existingDriverMap := make(map[string]bool)
	for _, driver := range existingDrivers {
		existingDriverMap[driver.DriverID] = true
	}

	// Filter out duplicates
	var uniqueDrivers []model.DriverLocation
	for _, driver := range newDrivers {
		if !existingDriverMap[driver.DriverID] {
			uniqueDrivers = append(uniqueDrivers, driver)
		}
	}

	return uniqueDrivers
}

// formatDriverResponse creates a formatted response from the matched drivers
func (s *matchingService) formatDriverResponse(user model.EnrichedUserLocation, drivers []model.DriverLocation) model.DriverResponse {
	response := model.DriverResponse{
		UserID:      user.UserID,
		RequestTime: time.Now().Unix(),
		Status:      "SUCCESS",
	}

	// If no drivers found, set appropriate status
	if len(drivers) == 0 {
		response.Status = "NO_DRIVERS_AVAILABLE"
		return response
	}

	// Format driver information
	driverInfos := make([]model.DriverInfo, len(drivers))
	for i, driver := range drivers {
		driverInfos[i] = model.DriverInfo{
			DriverID:    driver.DriverID,
			VehicleType: driver.VehicleType,
			Distance:    driver.Distance,
			ETA:         driver.ETA,
		}
	}

	response.Drivers = driverInfos
	return response
}

// processMatchingResults handles the results of driver matching
func (s *matchingService) processMatchingResults(user model.EnrichedUserLocation, drivers []model.DriverLocation) {
	if len(drivers) == 0 {
		log.Printf("No drivers available for user %s", user.UserID)
		return
	}

	log.Printf("Found %d drivers for user %s", len(drivers), user.UserID)

	response := s.formatDriverResponse(user, drivers)


    ctx := context.Background()
    userKey := fmt.Sprintf("user:%s:matches", user.UserID)
    
    responseJSON, err := json.Marshal(response)
    if err != nil {
        log.Printf("Error marshaling response to JSON: %v", err)
        return
    }
    
    // Store in Redis with 5 minute expiration
    err = s.redisClient.Set(ctx, userKey, responseJSON, 5*time.Minute).Err()
    if err != nil {
        log.Printf("Error storing matches in Redis: %v", err)
        return
    }
    
    log.Printf("Stored %d driver matches in Redis for user %s", len(drivers), user.UserID)
    
    notification := map[string]interface{}{
        "event": "driver_matches_updated",
        "user_id": user.UserID,
        "match_count": len(drivers),
        "timestamp": time.Now().Unix(),
    }
    
    notificationJSON, err := json.Marshal(notification)
    if err != nil {
        log.Printf("Error marshaling notification: %v", err)
        return
    }
    
    err = s.redisClient.Publish(ctx, "user_updates", notificationJSON).Err()
    if err != nil {
        log.Printf("Error publishing update notification: %v", err)
    }
}