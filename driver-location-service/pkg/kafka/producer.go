package kafka

import (
	"driver-location-service/internal/model"
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
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}
	log.Printf("Kafka producer created with brokers: %v", brokers)
	log.Printf("Kafka topic format: %s", topicFmt)
	return &Producer{
		producer: producer,
		topicFmt: topicFmt,
	}, nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

func (p *Producer) SendToProducer(loc model.Location) error {
	jsonData, err := json.Marshal(loc)
	if err != nil {
		return fmt.Errorf("marshaling error: %w", err)
	}

	topic := fmt.Sprintf(p.topicFmt, loc.City)
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(jsonData),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
