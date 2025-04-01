package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
)

// UserLocation represents a user's location data
type UserLocation struct {
	UserID      string  `json:"user_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	RequestType string  `json:"request_type"`
}

var (
	// Kafka broker addresses
	brokers = []string{"localhost:9101", "localhost:9102", "localhost:9103"}
	
	// Geographic boundaries for each city
	cityBounds = map[string][2][2]float64{
		"mumbai": {{18.96, 72.79}, {19.25, 72.98}},
		"pune":   {{18.45, 73.75}, {18.65, 74.00}},
		"delhi":  {{28.40, 76.80}, {28.90, 77.40}},
	}
		
	// User ID prefixes
	userPrefix = []string{"MUM", "PUN", "DEL"}
)

// createTopics creates Kafka topics if they don't exist
func createTopics(brokerList []string, topics []string, numPartitions int, replicationFactor int) error {
	config := sarama.NewConfig()
	config.Version = sarama.V3_5_0_0
	
	// Create admin client
	admin, err := sarama.NewClusterAdmin(brokerList, config)
	if err != nil {
		return fmt.Errorf("failed to create cluster admin: %v", err)
	}
	defer admin.Close()
	
	// Create each topic
	for _, topic := range topics {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     int32(numPartitions),
			ReplicationFactor: int16(replicationFactor),
			ConfigEntries: map[string]*string{
				"retention.ms": StringPtr("604800000"), // 7 days retention
			},
		}
		
		err := admin.CreateTopic(topic, topicDetail, false)
		if err != nil {
			// If topic already exists, just log and continue
			if err == sarama.ErrTopicAlreadyExists {
				fmt.Printf("Topic %s already exists\n", topic)
				continue
			}
			return fmt.Errorf("failed to create topic %s: %v", topic, err)
		}
		fmt.Printf("Created topic %s with %d partitions and replication factor %d\n", 
			topic, numPartitions, replicationFactor)
	}
	
	return nil
}

func main() {
	// Set random seed
	rand.Seed(time.Now().UnixNano())
	
	// Prepare topics list
	var topics []string
	for city := range cityBounds {
		topics = append(topics, fmt.Sprintf("%s-users", city))
	}
	
	// Create topics with 6 partitions and replication factor 1
	fmt.Println("Creating topics...")
	if err := createTopics(brokers, topics, 2, 1); err != nil {
		fmt.Printf("Warning: Topic creation issue: %v\n", err)
		// Continue execution instead of exiting
	}
	
	// Configure Sarama producer
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = 10 * time.Second
	config.Version = sarama.V3_5_0_0
	
	// Set batch configuration for better performance
	config.Producer.Flush.Frequency = 100 * time.Millisecond
	config.Producer.Flush.MaxMessages = 10

	// Create a new sync producer
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(fmt.Sprintf("Failed to create producer: %v", err))
	}
	defer producer.Close()

	// Maintain pool of generated user IDs
	userPool := generateUserPool(1000) 

	totalMessagesTransmitted := 0
	startTime := time.Now()
	
	// Setup graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	
	// Create ticker for message generation
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	// Create ticker for stats reporting
	statsTicker := time.NewTicker(5 * time.Second)
	defer statsTicker.Stop()
	
	fmt.Println("Starting to produce user location data...")
	
	// Main loop for sending messages
	running := true
	for running {
		select {
		case <-ticker.C:
			for city := range cityBounds {
				// Create location data for a random user
				loc := UserLocation{
					UserID:      userPool[rand.Intn(len(userPool))],
					City:        city,
					Latitude:    randomInRange(cityBounds[city][0][0], cityBounds[city][1][0]),
					Longitude:   randomInRange(cityBounds[city][0][1], cityBounds[city][1][1]),
					Timestamp:   time.Now().Unix(),
				}

				// Send the location to Kafka
				if err := sendToKafka(producer, loc); err != nil {
					fmt.Printf("Error sending to %s: %v\n", city, err)
					continue
				}
				
				totalMessagesTransmitted++
			}
			
		case <-statsTicker.C:
			elapsed := time.Since(startTime).Seconds()
			rate := float64(totalMessagesTransmitted) / elapsed
			fmt.Printf("Sent %d messages (%.2f msgs/sec)\n", totalMessagesTransmitted, rate)
			
		case sig := <-signals:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			running = false
		}
	}
	
	fmt.Println("Producer stopped successfully")
}

// generateUserPool creates a pool of random user IDs
func generateUserPool(size int) []string {
	pool := make([]string, size)
	for i := 0; i < size; i++ {
		pool[i] = fmt.Sprintf("%s-%04d%04d",
			userPrefix[rand.Intn(len(userPrefix))],
			rand.Intn(10000),
			rand.Intn(10000),
		)
	}
	return pool
}

// randomInRange returns a random float64 between min and max
func randomInRange(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// StringPtr returns a pointer to the string value passed in
func StringPtr(s string) *string {
	return &s
}

// sendToKafka sends the user location data to the appropriate Kafka topic
func sendToKafka(producer sarama.SyncProducer, loc UserLocation) error {
	jsonData, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: fmt.Sprintf("%s-users", loc.City),
		Key:   sarama.StringEncoder(loc.UserID), // Added key for partitioning
		Value: sarama.ByteEncoder(jsonData),
	}

	_, _, err = producer.SendMessage(msg)
	return err
}
