package main

import (
	"log"
)

const (
	KafkaBroker   = "localhost:9092"
	DriverTopic   = "driver-location-updates"
	RiderTopic    = "rider-location-updates"
	ServerAddress = ":6969"
)

func main() {
	brokers := []string{KafkaBroker}

	if err := createTopic(brokers, DriverTopic); err != nil {
		log.Fatalf("Failed to create driver topic: %v", err)
	}
	if err := createTopic(brokers, RiderTopic); err != nil {
		log.Fatalf("Failed to create rider topic: %v", err)
	}

	// Start HTTP server
	StartServer()
}
