package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/IBM/sarama"
)

type Location struct {
	DriverID    string  `json:"driver_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	VehicleType string  `json:"vehicle_type"`
	Status      string  `json:"status"`
}

var (
	brokers = []string{"localhost:9101", "localhost:9102", "localhost:9103"}
	cityBounds = map[string][2][2]float64{
		"mumbai": {{18.96, 72.79}, {19.25, 72.98}},
		"pune":   {{18.45, 73.75}, {18.65, 74.00}},
		"delhi":  {{28.40, 76.80}, {28.90, 77.40}},
	}
	vehicleTypes = []string{"STANDARD", "PREMIUM", "POOL", "LUXURY"}
	statuses     = []string{"ACTIVE", "INACTIVE", "BUSY"}
	driverPrefix = []string{"MH", "DL", "GA", "KA"}
)

func main() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = 10 * time.Second
	config.Version = sarama.V3_5_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
	defer producer.Close()

	rand.Seed(time.Now().UnixNano())
	
	// Maintain pool of generated driver IDs
	driverPool := generateDriverPool(1000) 

	totalMessagesTransmitted := 0

	for i := 0; ; i++ {
		for city := range cityBounds {
			loc := Location{
				DriverID:    driverPool[rand.Intn(len(driverPool))],
				City:        city,
				Latitude:    randomInRange(cityBounds[city][0][0], cityBounds[city][1][0]),
				Longitude:   randomInRange(cityBounds[city][0][1], cityBounds[city][1][1]),
				Timestamp:   time.Now().Unix(),
				VehicleType: vehicleTypes[rand.Intn(len(vehicleTypes))],
				Status:      weightedStatus(0.8), // 80% active
			}

			if err := sendToKafka(producer, loc); err != nil {
				fmt.Printf("Error sending to %s: %v\n", city, err)
				continue
			}
			fmt.Printf("Sent %+v\n", loc)
			totalMessagesTransmitted++
			fmt.Printf("%+v", totalMessagesTransmitted)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func generateDriverPool(size int) []string {
	pool := make([]string, size)
	for i := 0; i < size; i++ {
		pool[i] = fmt.Sprintf("%s-%04d%04d",
			driverPrefix[rand.Intn(len(driverPrefix))],
			rand.Intn(10000),
			rand.Intn(10000),
		)
	}
	return pool
}

func weightedStatus(activeProb float64) string {
	if rand.Float64() < activeProb {
		return "ACTIVE"
	}
	return statuses[rand.Intn(len(statuses)-1)+1]
}

func randomInRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
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
