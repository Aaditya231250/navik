package main

import (
	"log"
)

func main() {
	brokers := []string{"localhost:9092"}
	topic := "driver-position-update"

	if err := createTopic(brokers, topic); err != nil {
		log.Fatal(err)
		return
	}

	// Start HTTP server
	StartServer()
}
