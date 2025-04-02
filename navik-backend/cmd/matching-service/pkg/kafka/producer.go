package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"matching-service/internal/model"

	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topicFmt string
}

func NewProducer(brokers []string, topicFmt string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Net.DialTimeout = 10 * time.Second
	config.Version = sarama.V3_5_0_0

	// Set batch configuration for better performance
	config.Producer.Flush.Frequency = 100 * time.Millisecond
	config.Producer.Flush.MaxMessages = 10

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Error connecting to configured brokers: %v", err)
		log.Printf("Attempting to connect to localhost fallback...")

		fallbackBrokers := []string{"localhost:9092"}
		producer, err = sarama.NewSyncProducer(fallbackBrokers, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create producer with all broker options: %w", err)
		}
		log.Printf("Connected to fallback broker: %v", fallbackBrokers)
	} else {
		log.Printf("Successfully connected to brokers: %v", brokers)
	}

	log.Printf("Kafka topic format: %s", topicFmt)
	return &Producer{
		producer: producer,
		topicFmt: topicFmt,
	}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) SendLocation(loc model.UserLocation) error {
	jsonData, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	topic := fmt.Sprintf(p.topicFmt, loc.City)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(loc.UserID),
		Value: sarama.ByteEncoder(jsonData),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
