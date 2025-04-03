package kafka

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

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

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Error connecting to configured brokers: %v", err)
		log.Printf("Attempting to connect to localhost fallback...")

		fallbackBrokers := []string{"kafka-pune:29092"}
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

func (p *Producer) SendToProducer(data interface{}, topicKey string, messageKey string) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	topic := fmt.Sprintf(p.topicFmt, topicKey)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonData),
	}

	// Set the message key only if it's not empty
	if messageKey != "" {
		msg.Key = sarama.StringEncoder(messageKey)
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
