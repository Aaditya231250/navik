package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
	"location-service/internal/model"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topics   []string
	handler  func(model.Location) error
}

func NewConsumer(brokers []string, groupID string, topics []string, handler func(model.Location) error) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V3_5_0_0

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		topics:   topics,
		handler:  handler,
	}, nil
}

// Consume starts consuming messages from Kafka
func (c *Consumer) Consume(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Handle SIGINT and SIGTERM
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming in a goroutine
	go func() {
		for {
			if err := c.consumer.Consume(ctx, c.topics, c); err != nil {
				log.Printf("Error from consumer: %v", err)
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-signals
	log.Println("Shutting down consumer")
	return nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	return c.consumer.Close()
}

// Setup is run at the beginning of a new session
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim is called for each message
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var loc model.Location
		if err := json.Unmarshal(message.Value, &loc); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		if err := c.handler(loc); err != nil {
			log.Printf("Error handling message: %v", err)
			continue
		}

		session.MarkMessage(message, "")
	}
	return nil
}
