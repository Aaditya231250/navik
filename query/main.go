package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"sync"
	"math/rand"
    "sort"
    "strings"
	"strconv"
    "github.com/gorilla/websocket"
	"github.com/IBM/sarama"
	"github.com/uber/h3-go/v3"
	"github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/credentials"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/dynamodb"
)

// DriverNotification represents a notification to be sent to a driver
type DriverNotification struct {
	Type        string  `json:"type"`
	UserID      string  `json:"user_id"`
	DriverID    string  `json:"driver_id"`
	Priority    int     `json:"priority"`
	PickupLat   float64 `json:"pickup_lat"`
	PickupLong  float64 `json:"pickup_long"`
	Distance    float64 `json:"distance_km"`
	ETA         int     `json:"eta_minutes"`
	RequestTime int64   `json:"request_time"`
	ExpiresAt   int64   `json:"expires_at"`
}

// UserLocation represents a user's location data from Kafka
type UserLocation struct {
	UserID      string  `json:"user_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	RequestType string  `json:"request_type"`
}

// EnrichedUserLocation adds H3 indices to user location
type EnrichedUserLocation struct {
	UserLocation
	H3Index9 string
	H3Index8 string
	H3Index7 string
}

// DriverLocation represents a driver's location from the database
type DriverLocation struct {
    // Base driver information
    DriverID    string    `json:"driver_id" dynamodbav:"DriverID"`
    Latitude    float64   `json:"latitude"`
    Longitude   float64   `json:"longitude"`
    Location    string    `json:"location" dynamodbav:"Location"`
    VehicleType string    `json:"vehicle_type" dynamodbav:"Vehicle"`
    Status      string    `json:"status" dynamodbav:"Status"`
    LastUpdated time.Time `json:"last_updated"`
    
    // H3 indices at different resolutions
    H3Res9      string    `json:"h3_res9" dynamodbav:"H3Res9"`
    H3Res8      string    `json:"h3_res8" dynamodbav:"H3Res8"`
    H3Res7      string    `json:"h3_res7" dynamodbav:"H3Res7"`
    
    // DynamoDB keys
    // PK          string    `json:"-" dynamodbav:"PK"`
    // SK          string    `json:"-" dynamodbav:"SK"`
    // GSI1PK      string    `json:"-" dynamodbav:"GSI1PK"`
    // GSI1SK      string    `json:"-" dynamodbav:"GSI1SK"`
    // GSI2PK      string    `json:"-" dynamodbav:"GSI2PK"` // For H8 queries
    // GSI2SK      string    `json:"-" dynamodbav:"GSI2SK"`
    // GSI3PK      string    `json:"-" dynamodbav:"GSI3PK"` // For H7 queries
    // GSI3SK      string    `json:"-" dynamodbav:"GSI3SK"`
    
    // Time-related fields
    UpdatedAt   int64     `json:"-" dynamodbav:"UpdatedAt"`
    ExpiresAt   int64     `json:"-" dynamodbav:"ExpiresAt"`
    
    // Matching-related fields (not stored in DB)
    Distance    float64   `json:"distance,omitempty"`
    ETA         int       `json:"eta_minutes,omitempty"`
}

// DriverResponse represents the formatted response to send back to the user
type DriverResponse struct {
    UserID      string           `json:"user_id"`
    RequestTime int64            `json:"request_time"`
    Drivers     []DriverInfo     `json:"drivers"`
    Status      string           `json:"status"`
}

// DriverInfo contains the essential information about a matched driver
type DriverInfo struct {
    DriverID    string  `json:"driver_id"`
    VehicleType string  `json:"vehicle_type"`
    Distance    float64 `json:"distance_km"`
    ETA         int     `json:"eta_minutes"`
}



// Database connection
var (
    ddb               *dynamodb.DynamoDB
    // messagesReceived  atomic.Int64
    // messagesProcessed atomic.Int64
    // messagesFailedTotal atomic.Int64
    // ddbWriteAttempts  atomic.Int64
    // ddbWriteSuccesses atomic.Int64
    // ddbWriteFailures  atomic.Int64
    // batchSize         = 25
    // batchInterval     = 1 * time.Second
    // batchMutex        sync.Mutex
    // itemBatches       = make(map[string][]*dynamodb.WriteRequest)
    // batchTimers       = make(map[string]*time.Timer)
	MinDriversToReturn = 5
    MaxDistanceKm      = 10.0
    notificationConn *websocket.Conn
    notificationMutex sync.Mutex
    isConnected bool = false
)

// Initialize random seed
func init() {
    rand.Seed(time.Now().UnixNano())
}

func main() {
	dynamoEndpoint := "http://localhost:8000"

    sess := session.Must(session.NewSession(&aws.Config{
        Endpoint:    aws.String(dynamoEndpoint),
        Region:      aws.String("us-west-2"),
        Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
        DisableSSL:  aws.Bool(true),
    }))
    ddb = dynamodb.New(sess)

	// Setup Kafka configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_8_0_0 // Use appropriate version

	// Define Kafka brokers
	brokers := []string{"localhost:9101", "localhost:9102", "localhost:9103"}

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(brokers, "matching-service", config)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}
	defer consumerGroup.Close()

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create consumer handler
	handler := &ConsumerGroupHandler{}

	// Start consuming in a goroutine
	go func() {
		// List of topics to consume (only user topics)
		topics := []string{"mumbai-users", "pune-users", "delhi-users"}
		
		for {
			// Consume from topics
			if err := consumerGroup.Consume(ctx, topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				time.Sleep(5 * time.Second) // Wait before retrying
				continue
			}
			
			// Check if context was cancelled
			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Matching service started. Waiting for user requests...")

	// Wait for termination signal
	<-signals
	log.Println("Received termination signal. Shutting down...")
	cancel()
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler interface
type ConsumerGroupHandler struct {
    // Add fields for tracking resources or state if needed
    activeWorkers sync.WaitGroup
}

// Setup is called when the consumer group session is starting
func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
    log.Printf("Consumer group session started: %s", session.MemberID())
    // Initialize any session-specific resources here
    return nil
}

// Cleanup is called when the consumer group session is ending
func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
    log.Printf("Consumer group session ending: %s", session.MemberID())
    // Wait for any goroutines to finish if you're tracking them
    h.activeWorkers.Wait()
    // Clean up any session-specific resources here
    return nil
}

// ConsumeClaim processes messages from a partition claim
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
    // Track active workers if needed
    for message := range claim.Messages() {
        // For long-running tasks, you might want to check if context is done
        select {
        case <-session.Context().Done():
            return nil
        default:
            h.activeWorkers.Add(1)
            go func(msg *sarama.ConsumerMessage) {
                defer h.activeWorkers.Done()
                if err := processUserMessage(msg); err != nil {
                    log.Printf("Error processing message: %v", err)
                }
                session.MarkMessage(msg, "")
            }(message)
        }
    }
    return nil
}


// processUserMessage handles user location messages
func processUserMessage(message *sarama.ConsumerMessage) error {
	var userLoc UserLocation
	if err := json.Unmarshal(message.Value, &userLoc); err != nil {
		return fmt.Errorf("failed to unmarshal user location: %w", err)
	}
	
	// Enrich with H3 indices
	enrichedUser := enrichUserLocation(userLoc)
	
	log.Printf("Received user request: %s at H3-9: %s", 
		enrichedUser.UserID, enrichedUser.H3Index9)
	
	// Trigger the matching algorithm
	drivers, err := findDriversForUser(enrichedUser)
	if err != nil {
		return fmt.Errorf("error finding drivers: %w", err)
	}
	
	// Process the matching results
	processMatchingResults(enrichedUser, drivers)
	
	return nil
}

// enrichUserLocation adds H3 indices to a user location
func enrichUserLocation(loc UserLocation) EnrichedUserLocation {
	h3Index9 := geoToH3Index(loc.Latitude, loc.Longitude, 9)
	h3Index8 := geoToH3Index(loc.Latitude, loc.Longitude, 8)
	h3Index7 := geoToH3Index(loc.Latitude, loc.Longitude, 7)
	
	return EnrichedUserLocation{
		UserLocation: loc,
		H3Index9:     h3Index9,
		H3Index8:     h3Index8,
		H3Index7:     h3Index7,
	}
}

// geoToH3Index converts lat/lng to H3 index string
func geoToH3Index(lat, lng float64, resolution int) string {
	h3Index := h3.FromGeo(h3.GeoCoord{
		Latitude:  lat,
		Longitude: lng,
	}, resolution)
	
	return h3.ToString(h3Index)
}

// findDriversForUser implements the matching algorithm to find drivers for a user
func findDriversForUser(user EnrichedUserLocation) ([]DriverLocation, error) {
    // Step 1: Try to find drivers in the exact H9 cell
    drivers, err := findDriversInH9Cell(user.H3Index9)
    if err != nil {
        return nil, fmt.Errorf("error querying H9 cell: %w", err)
    }
    
    log.Printf("Found %d drivers in exact H9 cell %s", len(drivers), user.H3Index9)
    
    // If we found enough drivers, rank and return them
    if len(drivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(drivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }
    
    // Step 2: Not enough drivers, try H9 neighbors
    h9Neighbors := getH3Neighbors(user.H3Index9)
    log.Printf("Looking for drivers in %d neighboring H9 cells", len(h9Neighbors))
    
    // Query for drivers in neighboring H9 cells
    neighborDrivers, err := findDriversInH9Cells(h9Neighbors)
    if err != nil {
        return nil, fmt.Errorf("error querying H9 neighbor cells: %w", err)
    }
    
    // Combine with drivers from the exact cell
    allDrivers := append(drivers, neighborDrivers...)
    log.Printf("Found total of %d drivers in H9 cell and neighbors", len(allDrivers))
    
    // If we found enough drivers, rank and return them
    if len(allDrivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(allDrivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }
    
    // Step 3: Still not enough drivers, move up to H8 cell
    log.Printf("Not enough drivers in H9 cells, moving up to H8 cell %s", user.H3Index8)
    
    h8Drivers, err := findDriversInH8Cell(user.H3Index8)
    if err != nil {
        return nil, fmt.Errorf("error querying H8 cell: %w", err)
    }
    
    log.Printf("Found %d drivers in H8 cell", len(h8Drivers))
    
    // Filter out drivers we already found in H9 cells to avoid duplicates
    h8Drivers = filterOutDuplicateDrivers(h8Drivers, allDrivers)
    
    // Combine with drivers from H9 cells
    allDrivers = append(allDrivers, h8Drivers...)
    
    // If we found enough drivers, rank and return them
    if len(allDrivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(allDrivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }
    
    // Step 4: Still not enough drivers, try H8 neighbors
    h8Neighbors := getH3Neighbors(user.H3Index8)
    log.Printf("Looking for drivers in %d neighboring H8 cells", len(h8Neighbors))
    
    // Query for drivers in neighboring H8 cells
    h8NeighborDrivers, err := findDriversInH8Cells(h8Neighbors)
    if err != nil {
        return nil, fmt.Errorf("error querying H8 neighbor cells: %w", err)
    }
    
    // Filter out drivers we already found
    h8NeighborDrivers = filterOutDuplicateDrivers(h8NeighborDrivers, allDrivers)
    
    // Combine with all previous drivers
    allDrivers = append(allDrivers, h8NeighborDrivers...)
    log.Printf("Found total of %d drivers after H8 neighbors", len(allDrivers))
    
    // If we found enough drivers, rank and return them
    if len(allDrivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(allDrivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }
    
    // Step 5: Still not enough drivers, move up to H7 cell
    log.Printf("Not enough drivers in H8 cells, moving up to H7 cell %s", user.H3Index7)
    
    h7Drivers, err := findDriversInH7Cell(user.H3Index7)
    if err != nil {
        return nil, fmt.Errorf("error querying H7 cell: %w", err)
    }
    
    log.Printf("Found %d drivers in H7 cell", len(h7Drivers))
    
    // Filter out drivers we already found to avoid duplicates
    h7Drivers = filterOutDuplicateDrivers(h7Drivers, allDrivers)
    
    // Combine with all previous drivers
    allDrivers = append(allDrivers, h7Drivers...)
    
    // If we found enough drivers, rank and return them
    if len(allDrivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(allDrivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }

	// Step 6: Still not enough drivers, try H7 neighbors
    h7Neighbors := getH3Neighbors(user.H3Index7)
    log.Printf("Looking for drivers in %d neighboring H7 cells", len(h7Neighbors))
    
    // Query for drivers in neighboring H7 cells
    h7NeighborDrivers, err := findDriversInH7Cells(h7Neighbors)
    if err != nil {
        return nil, fmt.Errorf("error querying H7 neighbor cells: %w", err)
    }
    
    // Filter out drivers we already found
    h7NeighborDrivers = filterOutDuplicateDrivers(h7NeighborDrivers, allDrivers)
    
    // Combine with all previous drivers
    allDrivers = append(allDrivers, h7NeighborDrivers...)
    log.Printf("Found total of %d drivers after H7 neighbors", len(allDrivers))
    
    // If we found enough drivers, rank and return them
    if len(allDrivers) >= MinDriversToReturn {
        // Assign random distances and rank drivers
        rankedDrivers := assignRandomDistances(allDrivers)
        return getTopDrivers(rankedDrivers, MinDriversToReturn), nil
    }
    
    // Continue to next phase if still not enough drivers
    return allDrivers, fmt.Errorf("no available drivers found in the vicinity")

}

// assignRandomDistances assigns random distances to drivers and sorts them
// This is a placeholder for the Google Maps API distance calculation
func assignRandomDistances(drivers []DriverLocation) []DriverLocation {
    // Assign random distances
    for i := range drivers {
        // Random distance between 0.5 and 5 km
        // H9 cells get shorter distances (0.5-2 km)
        // H8 cells get medium distances (1-3 km)
        // H7 cells get longer distances (2-5 km)
        drivers[i].Distance = 0.5 + rand.Float64()*4.5
    }
    
    // Sort by distance
    sort.Slice(drivers, func(i, j int) bool {
        return drivers[i].Distance < drivers[j].Distance
    })
    
    return drivers
}

func parseInt64(s string) int64 {
    val, _ := strconv.ParseInt(s, 10, 64)
    return val
}

func safePrefix(s string, length int) string {
    if len(s) < length {
        return s
    }
    return s[:length]
}

// findDriversInH9Cell queries the database for drivers in a specific H9 cell
func findDriversInH9Cell(h3Index string) ([]DriverLocation, error) {
    h3Prefix := safePrefix(h3Index, 5)
    gsi1pk := fmt.Sprintf("ACTIVE#H3#9#%s", h3Prefix)
    log.Printf("Querying with GSI1PK: %s", gsi1pk)
    input := &dynamodb.QueryInput{
        TableName: aws.String("DriverLocations"),
        IndexName: aws.String("StatusH3Index"), // Use the Global Secondary Index
        KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":gsi1pk": {
                S: aws.String(fmt.Sprintf("ACTIVE#H3#9#%s", h3Prefix)),
            },
        },
    }
    
    // Execute the query
    result, err := ddb.Query(input)
    if err != nil {
        return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
    }
    
    // If the result is nil, return an empty slice
    if result == nil {
        return []DriverLocation{}, nil
    }
    
    var drivers []DriverLocation
    for _, item := range result.Items {
        // Parse location string into latitude and longitude
        if item["location"] == nil || item["location"].S == nil {
            log.Printf("Warning: Item missing location field: %v", item)
            continue
        }
        
        locString := *item["location"].S
        locParts := strings.Split(locString, ",")

        if len(locParts) != 2 {
            log.Printf("Warning: Invalid location format: %s", locString)
            continue
        }
        
        lat, err := strconv.ParseFloat(locParts[0], 64)
        if err != nil {
            log.Printf("Warning: Failed to parse latitude: %s", locParts[0])
            continue
        }
        
        lng, err := strconv.ParseFloat(locParts[1], 64)
        if err != nil {
            log.Printf("Warning: Failed to parse longitude: %s", locParts[1])
            continue
        }
        
        // Check for required fields before accessing them
        if item["driver_id"] == nil || item["driver_id"].S == nil ||
           item["h3_res9"] == nil || item["h3_res9"].S == nil ||
           item["status"] == nil || item["status"].S == nil ||
           item["vehicle_type"] == nil || item["vehicle_type"].S == nil ||
           item["updated_at"] == nil || item["updated_at"].N == nil {
            log.Printf("Warning: Item missing required fields: %v", item)
            continue
        }
        
        driver := DriverLocation{
            DriverID:    *item["driver_id"].S,
            Latitude:    lat,
            Longitude:   lng,
            H3Res9:      *item["h3_res9"].S,
            Status:      *item["status"].S,
            VehicleType: *item["vehicle_type"].S,
            LastUpdated: time.Unix(parseInt64(*item["updated_at"].N), 0),
        }
        drivers = append(drivers, driver)
    }
    
    log.Printf("Found %d drivers in H9 cell %s", len(drivers), h3Index)
    return drivers, nil
}

// findDriversInH9Cells queries DynamoDB for drivers in multiple H9 cells
func findDriversInH9Cells(h3Indices []string) ([]DriverLocation, error) {
    if len(h3Indices) == 0 {
        return []DriverLocation{}, nil
    }
    
    // Use goroutines to query each H9 cell in parallel
    var wg sync.WaitGroup
    var mu sync.Mutex
    var allDrivers []DriverLocation
    var queryErrors []error
    
    for _, h3Index := range h3Indices {
        wg.Add(1)
        go func(index string) {
            defer wg.Done()
            
            drivers, err := findDriversInH9Cell(index)
            if err != nil {
                mu.Lock()
                queryErrors = append(queryErrors, fmt.Errorf("failed to query H9 cell %s: %w", index, err))
                mu.Unlock()
                return
            }
            
            mu.Lock()
            allDrivers = append(allDrivers, drivers...)
            mu.Unlock()
        }(h3Index)
    }
    
    wg.Wait()
    
    if len(queryErrors) > 0 {
        return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
    }
    
    return allDrivers, nil
}

// findDriversInH8Cell queries DynamoDB for drivers in a specific H8 cell
func findDriversInH8Cell(h8Index string) ([]DriverLocation, error) {
    h8Prefix := safePrefix(h8Index, 5)
    
    input := &dynamodb.QueryInput{
        TableName: aws.String("DriverLocations"),
        IndexName: aws.String("StatusH3Res8Index"), // Use the Global Secondary Index for H8
        KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":gsi2pk": {
                S: aws.String(fmt.Sprintf("ACTIVE#H3#8#%s", h8Prefix)),
            },
        },
    }
    
    result, err := ddb.Query(input)
    if err != nil {
        return nil, fmt.Errorf("failed to query DynamoDB for H8 cell %s: %w", h8Index, err)
    }
    
    var drivers []DriverLocation
    for _, item := range result.Items {
        // Check if required fields exist before accessing them
        if item["location"] == nil || item["location"].S == nil {
            log.Printf("Warning: Item missing location field: %v", item)
            continue
        }
        
        // Parse location string into latitude and longitude
        locString := *item["location"].S
        locParts := strings.Split(locString, ",")
        
        if len(locParts) != 2 {
            log.Printf("Warning: Invalid location format: %s", locString)
            continue
        }
        
        lat, err := strconv.ParseFloat(locParts[0], 64)
        if err != nil {
            log.Printf("Warning: Failed to parse latitude: %s", locParts[0])
            continue
        }
        
        lng, err := strconv.ParseFloat(locParts[1], 64)
        if err != nil {
            log.Printf("Warning: Failed to parse longitude: %s", locParts[1])
            continue
        }
        
        // Check for required fields before accessing them
        if item["driver_id"] == nil || item["driver_id"].S == nil ||
           item["vehicle_type"] == nil || item["vehicle_type"].S == nil ||
           item["status"] == nil || item["status"].S == nil ||
           item["updated_at"] == nil || item["updated_at"].N == nil {
            log.Printf("Warning: Item missing required fields: %v", item)
            continue
        }
        
        driver := DriverLocation{
            DriverID:    *item["driver_id"].S,
            Latitude:    lat,
            Longitude:   lng,
            VehicleType: *item["vehicle_type"].S,
            Status:      *item["status"].S,
            LastUpdated: time.Unix(parseInt64(*item["updated_at"].N), 0),
        }
        drivers = append(drivers, driver)
    }
    
    log.Printf("Found %d drivers in H8 cell %s", len(drivers), h8Index)
    return drivers, nil
}

// findDriversInH8Cells queries DynamoDB for drivers in multiple H8 cells
func findDriversInH8Cells(h8Indices []string) ([]DriverLocation, error) {
    if len(h8Indices) == 0 {
        return []DriverLocation{}, nil
    }
    
    // Use goroutines to query each H8 cell in parallel
    var wg sync.WaitGroup
    var mu sync.Mutex
    var allDrivers []DriverLocation
    var queryErrors []error
    
    for _, h8Index := range h8Indices {
        wg.Add(1)
        go func(index string) {
            defer wg.Done()
            
            drivers, err := findDriversInH8Cell(index)
            if err != nil {
                mu.Lock()
                queryErrors = append(queryErrors, fmt.Errorf("failed to query H8 cell %s: %w", index, err))
                mu.Unlock()
                return
            }
            
            mu.Lock()
            allDrivers = append(allDrivers, drivers...)
            mu.Unlock()
        }(h8Index)
    }
    
    wg.Wait()
    
    if len(queryErrors) > 0 {
        return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
    }
    
    return allDrivers, nil
}

// findDriversInH7Cell queries DynamoDB for drivers in a specific H7 cell
func findDriversInH7Cell(h7Index string) ([]DriverLocation, error) {
    h7Prefix := safePrefix(h7Index, 5)
    
    input := &dynamodb.QueryInput{
        TableName: aws.String("DriverLocations"),
        IndexName: aws.String("StatusH3Res7Index"), // Use the Global Secondary Index for H7
        KeyConditionExpression: aws.String("GSI3PK = :gsi3pk"),
        ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
            ":gsi3pk": {
                S: aws.String(fmt.Sprintf("ACTIVE#H3#7#%s", h7Prefix)),
            },
        },
    }
    
    result, err := ddb.Query(input)
    if err != nil {
        return nil, fmt.Errorf("failed to query DynamoDB for H7 cell %s: %w", h7Index, err)
    }
    
    var drivers []DriverLocation
    for _, item := range result.Items {
        // Parse location string into latitude and longitude
        locParts := strings.Split(*item["Location"].S, ",")
        lat, _ := strconv.ParseFloat(locParts[0], 64)
        lng, _ := strconv.ParseFloat(locParts[1], 64)
        
        driver := DriverLocation{
            DriverID:    *item["DriverID"].S,
            Latitude:    lat,
            Longitude:   lng,
            VehicleType: *item["Vehicle"].S,
            Status:      *item["Status"].S,
            LastUpdated: time.Unix(parseInt64(*item["UpdatedAt"].N), 0),
        }
        drivers = append(drivers, driver)
    }
    
    return drivers, nil
}

// findDriversInH7Cells queries DynamoDB for drivers in multiple H7 cells
func findDriversInH7Cells(h7Indices []string) ([]DriverLocation, error) {
	if len(h7Indices) == 0 {
	return []DriverLocation{}, nil
	}
	// Use goroutines to query each H7 cell in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allDrivers []DriverLocation
	var queryErrors []error

	for _, h7Index := range h7Indices {
		wg.Add(1)
		go func(index string) {
			defer wg.Done()
			
			drivers, err := findDriversInH7Cell(index)
			if err != nil {
				mu.Lock()
				queryErrors = append(queryErrors, fmt.Errorf("failed to query H7 cell %s: %w", index, err))
				mu.Unlock()
				return
			}
			
			mu.Lock()
			allDrivers = append(allDrivers, drivers...)
			mu.Unlock()
		}(h7Index)
	}

	wg.Wait()

	if len(queryErrors) > 0 {
		return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
	}

	return allDrivers, nil
}

// getH3Neighbors returns the neighboring H3 cells for a given H3 index at any resolution
func getH3Neighbors(h3Index string) []string {
    // Convert string index to H3 index
    h3IndexInt := h3.FromString(h3Index)
    
    // Get the k-ring of neighbors (k=1 means immediate neighbors)
    neighbors := h3.KRing(h3IndexInt, 1)
    
    // Convert to strings and filter out the center cell (which is the original cell)
    var neighborStrings []string
    for _, neighbor := range neighbors {
        neighborString := h3.ToString(neighbor)
        if neighborString != h3Index {
            neighborStrings = append(neighborStrings, neighborString)
        }
    }
    
    return neighborStrings
}

// getTopDrivers returns the top N drivers from a ranked list
func getTopDrivers(drivers []DriverLocation, n int) []DriverLocation {
    if len(drivers) <= n {
        return drivers
    }
    return drivers[:n]
}

// filterOutDuplicateDrivers removes drivers that are already in the existingDrivers list
func filterOutDuplicateDrivers(newDrivers, existingDrivers []DriverLocation) []DriverLocation {
    // Create a map of existing driver IDs for quick lookup
    existingDriverMap := make(map[string]bool)
    for _, driver := range existingDrivers {
        existingDriverMap[driver.DriverID] = true
    }
    
    // Filter out duplicates
    var uniqueDrivers []DriverLocation
    for _, driver := range newDrivers {
        if !existingDriverMap[driver.DriverID] {
            uniqueDrivers = append(uniqueDrivers, driver)
        }
    }
    
    return uniqueDrivers
}

// formatDriverResponse creates a formatted response from the matched drivers
func formatDriverResponse(user EnrichedUserLocation, drivers []DriverLocation) DriverResponse {
    response := DriverResponse{
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
    driverInfos := make([]DriverInfo, len(drivers))
    for i, driver := range drivers {
        driverInfos[i] = DriverInfo{
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
func processMatchingResults(user EnrichedUserLocation, drivers []DriverLocation) {
    if len(drivers) == 0 {
        log.Printf("No drivers available for user %s", user.UserID)
        return
    }
    
    log.Printf("Found %d drivers for user %s", len(drivers), user.UserID)
    
    // Format response for the user
    response := formatDriverResponse(user, drivers)

    if(response.Status == "NO_DRIVERS_AVAILABLE") {
        log.Printf("No drivers available for user %s", user.UserID)
        return
    }

    // Ensure we're connected to notification service
    if err := connectToNotificationService(); err != nil {
        log.Printf("Failed to connect to notification service: %v", err)
        return
    }
    
    // Prepare driver notifications
    notifications := make([]DriverNotification, len(drivers))
    for i, driver := range drivers {
        notifications[i] = DriverNotification{
            Type:        "RIDE_REQUEST",
            UserID:      user.UserID,
            DriverID:    driver.DriverID,
            Priority:    i + 1, // Priority based on ranking
            PickupLat:   user.Latitude,
            PickupLong:  user.Longitude,
            Distance:    driver.Distance,
            ETA:         driver.ETA,
            RequestTime: time.Now().Unix(),
            ExpiresAt:   time.Now().Add(30 * time.Second).Unix(), // Driver has 30s to respond
        }
    }
    
    // Send notifications to notification service
    notificationMutex.Lock()
    defer notificationMutex.Unlock()
    
    if !isConnected {
        log.Printf("Not connected to notification service")
        return
    }
    
    notificationJSON, err := json.Marshal(notifications)
    if err != nil {
        log.Printf("Error marshaling notifications: %v", err)
        return
    }
    
    if err := notificationConn.WriteMessage(websocket.TextMessage, notificationJSON); err != nil {
        log.Printf("Error sending notifications: %v", err)
        isConnected = false
        return
    }
    
    log.Printf("Sent %d driver notifications for user %s", len(notifications), user.UserID)
}

func connectToNotificationService() error {
    notificationMutex.Lock()
    defer notificationMutex.Unlock()
    
    if isConnected {
        return nil
    }
    
    dialer := websocket.Dialer{
        HandshakeTimeout: 5 * time.Second,
    }
    
    conn, _, err := dialer.Dial("ws://notification-service:9080/ws/matching", nil)
    if err != nil {
        return fmt.Errorf("failed to connect to notification service: %w", err)
    }
    
    notificationConn = conn
    isConnected = true
    
    // Start a goroutine to handle connection lifecycle
    go func() {
        defer func() {
            notificationMutex.Lock()
            notificationConn.Close()
            notificationConn = nil
            isConnected = false
            notificationMutex.Unlock()
        }()
        
        for {
            // Read messages (acknowledgments, etc.)
            _, _, err := conn.ReadMessage()
            if err != nil {
                log.Printf("Notification service connection closed: %v", err)
                break
            }
        }
        
        // Attempt to reconnect after a delay
        time.Sleep(5 * time.Second)
        if err := connectToNotificationService(); err != nil {
            log.Printf("Failed to reconnect to notification service: %v", err)
        }
    }()
    
    log.Println("Connected to notification service")
    return nil
}
