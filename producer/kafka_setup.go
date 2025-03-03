package main

import (
	"fmt"
	"log"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"
)

// createTopic creates a Kafka topic with the given name
func createTopic(brokers []string, topic string) error {
	conn, err := kafka.Dial("tcp", brokers[0])
	if err != nil {
		return fmt.Errorf("failed to dial leader: %v", err)
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return fmt.Errorf("failed to get controller: %v", err)
	}

	controllerConn, err := kafka.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return fmt.Errorf("failed to dial controller: %v", err)
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return fmt.Errorf("failed to create topic: %v", err)
	}

	log.Printf("Topic '%s' created successfully", topic)
	return nil
}
