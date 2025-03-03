package main

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

// publishMessage sends a message to the Kafka topic
func publishMessage(brokers []string, topic, key, value string) error {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	ctx := context.Background()
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(key),
		Value: []byte(value),
	})

	if err != nil {
		return fmt.Errorf("failed to write message: %v", err)
	}

	log.Printf("Message published to topic '%s': %s", topic, value)
	return nil
}
