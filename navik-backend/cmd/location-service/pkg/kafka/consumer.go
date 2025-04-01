package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"location-service/internal/model"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type MessageHandler func(loc model.Location) error

type Consumer struct {
	consumers         []*kafka.Consumer
	handler           MessageHandler
	messagesReceived  *int64
	messagesProcessed *int64
	messagesFailed    *int64
}

type ClusterConfig struct {
	Name             string
	BootstrapServers string
	Topic            string
}

func NewConsumer(groupID string, clusters []ClusterConfig, handler MessageHandler,
	messagesReceived, messagesProcessed, messagesFailed *int64) (*Consumer, error) {
	consumers := make([]*kafka.Consumer, 0, len(clusters))

	for _, cluster := range clusters {
		config := &kafka.ConfigMap{
			"bootstrap.servers":  cluster.BootstrapServers,
			"group.id":           groupID,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": "false",
		}

		consumer, err := kafka.NewConsumer(config)
		if err != nil {
			// Clean up any previously created consumers
			for _, c := range consumers {
				c.Close()
			}
			return nil, fmt.Errorf("failed to create consumer for %s: %w", cluster.Name, err)
		}

		err = consumer.SubscribeTopics([]string{cluster.Topic}, nil)
		if err != nil {
			consumer.Close()
			// Clean up any previously created consumers
			for _, c := range consumers {
				c.Close()
			}
			return nil, fmt.Errorf("failed to subscribe to topic for %s: %w", cluster.Name, err)
		}

		log.Printf("[%s] Started consumer for topic %s", cluster.Name, cluster.Topic)
		consumers = append(consumers, consumer)
	}

	return &Consumer{
		consumers:         consumers,
		handler:           handler,
		messagesReceived:  messagesReceived,
		messagesProcessed: messagesProcessed,
		messagesFailed:    messagesFailed,
	}, nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	var wg sync.WaitGroup

	for i, consumer := range c.consumers {
		wg.Add(1)
		go func(id int, consumer *kafka.Consumer) {
			defer wg.Done()
			c.consumeMessages(ctx, id, consumer)
		}(i, consumer)
	}

	wg.Wait()
	return nil
}

func (c *Consumer) consumeMessages(ctx context.Context, id int, consumer *kafka.Consumer) {
	log.Printf("Starting consumer #%d", id)
	defer log.Printf("Stopping consumer #%d", id)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() != kafka.ErrTimedOut {
					log.Printf("Consumer #%d error: %v", id, err)
				}
				continue
			}

			atomic.AddInt64(c.messagesReceived, 1)

			if err := c.processMessage(msg); err == nil {
				consumer.CommitMessage(msg)
				atomic.AddInt64(c.messagesProcessed, 1)
				log.Printf("Consumer #%d processed message from partition %d, offset %d",
					id, msg.TopicPartition.Partition, msg.TopicPartition.Offset)
			} else {
				atomic.AddInt64(c.messagesFailed, 1)
				log.Printf("Consumer #%d error processing message from partition %d, offset %d: %v",
					id, msg.TopicPartition.Partition, msg.TopicPartition.Offset, err)

				// Commit anyway to avoid getting stuck
				consumer.CommitMessage(msg)
			}
		}
	}
}

func (c *Consumer) processMessage(msg *kafka.Message) error {
	var loc model.Location

	if err := json.Unmarshal(msg.Value, &loc); err != nil {
		log.Printf("Error parsing message: %v. Raw message: %s", err, string(msg.Value))
		return fmt.Errorf("failed to unmarshal location: %w", err)
	}

	if err := loc.Validate(); err != nil {
		return fmt.Errorf("invalid location data: %w", err)
	}

	return c.handler(loc)
}

func (c *Consumer) Close() {
	for _, consumer := range c.consumers {
		consumer.Close()
	}
}
