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
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	repo := repository.NewInMemoryLocationRepository()

	locationService := service.NewLocationService(repo, nil)

	locationHandler := func(loc model.Location) error {
		log.Printf("Processing location update for driver %s in %s", loc.DriverID, loc.City)
		return locationService.ProcessLocationUpdate(loc)
	}

	consumer, err := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.GroupID, cfg.Kafka.Topics, locationHandler)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting Kafka consumer...")
		if err := consumer.Consume(ctx); err != nil {
			log.Fatalf("Error in consumer: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down consumer...")
	cancel() // Cancel the consumer context
	log.Println("Consumer stopped gracefully")
}
