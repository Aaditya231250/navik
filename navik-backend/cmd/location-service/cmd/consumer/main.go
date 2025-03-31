// cmd/consumer/main.go
package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"location-service/internal/config"
	"location-service/internal/model"
	"location-service/internal/repository"
	"location-service/internal/service"
	"location-service/pkg/kafka"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize in-memory repository
	repo := repository.NewInMemoryLocationRepository()

	// Set up service
	locationService := service.NewLocationService(repo, nil) // No producer needed for consumer

	// Handler function for location updates
	locationHandler := func(loc model.Location) error {
		log.Printf("Processing location update for driver %s in %s", loc.DriverID, loc.City)
		return locationService.ProcessLocationUpdate(loc)
	}

	// Create and start consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topics, locationHandler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	// Context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consuming in a separate goroutine
	go func() {
		log.Println("Starting Kafka consumer...")
		if err := consumer.Consume(ctx); err != nil {
			log.Fatalf("Error in consumer: %v", err)
		}
	}()

	// Wait for termination signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down consumer...")
	cancel() // Cancel the consumer context
	log.Println("Consumer stopped gracefully")
}
