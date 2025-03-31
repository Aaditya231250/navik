package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/uber/h3-go/v4"
)

const (
	outputFile    = "driver_queries.log"
	h3Resolution  = 9 // Using H3 resolution 9 (~350m precision)
)

var cityReaders = map[string]struct {
	Topic  string
	Broker string
}{
	"mumbai": {"mumbai-drivers", "localhost:9101"},
	"pune":   {"pune-drivers", "localhost:9102"},
	"delhi":  {"delhi-drivers", "localhost:9103"},
}

type ProcessedRecord struct {
	DriverID    string    `json:"driver_id"`
	H3Index     string    `json:"h3_index"`
	City        string    `json:"city"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	VehicleType string    `json:"vehicle_type"`
	Timestamp   time.Time `json:"timestamp"`
}

func main() {
	// Setup output file with thread-safe writer
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to open output file: %v", err))
	}
	defer file.Close()

	var fileLock sync.Mutex
	logEntry := func(entry ProcessedRecord) {
		data, _ := json.Marshal(entry)
		fileLock.Lock()
		defer fileLock.Unlock()
		file.WriteString(string(data) + "\n")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup

	// Start consumers for each city
	for city, cfg := range cityReaders {
		wg.Add(1)
		go func(city string, cfg struct{ Topic, Broker string }) {
			defer wg.Done()

			reader := kafka.NewReader(kafka.ReaderConfig{
				Brokers:   []string{cfg.Broker},
				Topic:     cfg.Topic,
				Partition: 0,
				MinBytes:  10e3, // 10KB
				MaxBytes:  10e6, // 10MB
				MaxWait:   1 * time.Second,
			})
			defer reader.Close()

			fmt.Printf("Starting consumer for %s...\n", city)
			
			for {
				select {
				case <-ctx.Done():
					return
				default:
					msg, err := reader.FetchMessage(ctx)
					if err != nil {
						if ctx.Err() == nil {
							fmt.Printf("Error reading message: %v\n", err)
						}
						continue
					}

					// Process message
					var loc DriverLocation
					if err := json.Unmarshal(msg.Value, &loc); err != nil {
						fmt.Printf("Error decoding message: %v\n", err)
						continue
					}

					// Generate H3 index
					geo := h3.NewLatLng(loc.Latitude, loc.Longitude)
					index := h3.FromLatLng(geo, h3Resolution)

					// Store processed record
					logEntry(ProcessedRecord{
						DriverID:    loc.DriverID,
						H3Index:     index.String(),
						City:        city,
						Latitude:    loc.Latitude,
						Longitude:   loc.Longitude,
						VehicleType: loc.VehicleType,
						Timestamp:   time.Now().UTC(),
					})

					// Commit offset
					if err := reader.CommitMessages(ctx, msg); err != nil {
						fmt.Printf("Failed to commit offset: %v\n", err)
					}
				}
			}
		}(city, cfg)
	}

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nInitiating shutdown...")
	cancel()
	wg.Wait()
	fmt.Println("Consumer stopped successfully")
}

// Reuse DriverLocation struct from producer
type DriverLocation struct {
	DriverID    string  `json:"driver_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Accuracy    float64 `json:"accuracy"`
	Timestamp   int64   `json:"timestamp"`
	VehicleType string  `json:"vehicle_type"`
}
