package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

// City configurations
var cityConfigs = map[string]struct {
	Topic         string
	MinLat        float64
	MaxLat        float64
	MinLon        float64
	MaxLon        float64
	Broker        string
}{
	"mumbai": {
		Topic:  "mumbai-drivers",
		MinLat: 18.9,
		MaxLat: 19.2,
		MinLon: 72.7,
		MaxLon: 73.2,
		Broker: "localhost:9101",
	},
	"pune": {
		Topic:  "pune-drivers",
		MinLat: 18.4,
		MaxLat: 18.6,
		MinLon: 73.7,
		MaxLon: 74.0,
		Broker: "localhost:9102",
	},
	"delhi": {
		Topic:  "delhi-drivers",
		MinLat: 28.4,
		MaxLat: 28.9,
		MinLon: 76.8,
		MaxLon: 77.4,
		Broker: "localhost:9103",
	},
}

type DriverLocation struct {
	DriverID    string  `json:"driver_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Accuracy    float64 `json:"accuracy"`
	Timestamp   int64   `json:"timestamp"`
	VehicleType string  `json:"vehicle_type"`
}

func createTopicIfNotExist(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return fmt.Errorf("failed to dial broker %s: %w", broker, err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %w", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to controller: %w", err)
	}
	defer controllerConn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return fmt.Errorf("failed to read partitions: %w", err)
	}

	topicExists := false
	for _, p := range partitions {
		if p.Topic == topic {
			topicExists = true
			break
		}
	}

	if !topicExists {
		err = controllerConn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     2,
			ReplicationFactor: 1,
			ConfigEntries: []kafka.ConfigEntry{
				{ConfigName: "retention.ms", ConfigValue: "604800000"},
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create topic %s: %w", topic, err)
		}
		fmt.Printf("Created topic %s on broker %s\n", topic, broker)
	}
	return nil
}

func main() {
	// Create topics if they don't exist
	for city, cfg := range cityConfigs {
		if err := createTopicIfNotExist(cfg.Broker, cfg.Topic); err != nil {
			panic(fmt.Sprintf("Topic creation failed for %s: %v", city, err))
		}
	}

	// Initialize Kafka writers
	writers := make(map[string]*kafka.Writer)
	for city, cfg := range cityConfigs {
		writers[city] = &kafka.Writer{
			Addr:         kafka.TCP(cfg.Broker),
			Topic:        cfg.Topic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
			Transport: &kafka.Transport{
				DialTimeout: 10 * time.Second,
			},
		}
		defer writers[city].Close()
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Generate driver locations
	rand.Seed(time.Now().UnixNano())
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

productionLoop:
	for {
		select {
		case <-ticker.C:
			for city, cfg := range cityConfigs {
				lat := cfg.MinLat + rand.Float64()*(cfg.MaxLat-cfg.MinLat)
				lon := cfg.MinLon + rand.Float64()*(cfg.MaxLon-cfg.MinLon)

				msg := DriverLocation{
					DriverID:    fmt.Sprintf("DRV-%s-%04d", city, rand.Intn(10000)),
					City:        city,
					Latitude:    lat,
					Longitude:   lon,
					Accuracy:    5 + rand.Float64()*15,
					Timestamp:   time.Now().UnixMilli(),
					VehicleType: []string{"sedan", "hatchback", "suv"}[rand.Intn(3)],
				}

				value, _ := json.Marshal(msg)
				
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				err := writers[city].WriteMessages(ctx, kafka.Message{
					Key:   []byte(msg.DriverID),
					Value: value,
				})

				if err != nil {
					fmt.Printf("Failed to write to %s: %v\n", cfg.Topic, err)
				} else {
					fmt.Printf("Sent to %s: %s\n", cfg.Topic, msg.DriverID)
				}
			}

		case <-sigChan:
			fmt.Println("\nInitiating graceful shutdown...")
			break productionLoop
		}
	}

	fmt.Println("Producer stopped successfully")
}
