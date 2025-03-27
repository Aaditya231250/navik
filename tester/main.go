package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

type Location struct {
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

// Use all brokers in a single list
var brokers = []string{
	"localhost:9101", // Mumbai
	"localhost:9102", // Pune
	"localhost:9103", // Delhi
}

var cityBounds = map[string][2][2]float64{
	"mumbai": {{18.96, 72.79}, {19.25, 72.98}},
	"pune":   {{18.45, 73.75}, {18.65, 74.00}},
	"delhi":  {{28.40, 76.80}, {28.90, 77.40}},
}

func main() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = 10 * time.Second
	config.Version = sarama.V3_5_0_0 // Match Kafka 7.5.1 version

	// Create a single producer for all brokers
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
	defer producer.Close()

	rand.Seed(time.Now().UnixNano())
	for i := 0; ; i++ {
		for city := range cityBounds {
			loc := generateLocation(city)
			if err := sendToKafka(producer, loc); err != nil {
				fmt.Printf("Error sending to %s: %v\n", city, err)
				continue
			}
			fmt.Printf("Sent to %s: %+v\n", city, loc)
		}
		time.Sleep(1000000 * time.Microsecond)
	}
}

func generateLocation(city string) Location {
	bounds := cityBounds[city]
	return Location{
		City:      city,
		Latitude:  bounds[0][0] + rand.Float64()*(bounds[1][0]-bounds[0][0]),
		Longitude: bounds[0][1] + rand.Float64()*(bounds[1][1]-bounds[0][1]),
		Timestamp: time.Now().UnixNano(),
	}
}

func sendToKafka(producer sarama.SyncProducer, loc Location) error {
	jsonData, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s-locations", loc.City),
		Value: sarama.ByteEncoder(jsonData),
	}

	_, _, err = producer.SendMessage(msg)
	return err
}
