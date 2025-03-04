package main

import (
	"log"
	"github.com/IBM/sarama"
)

func publishMessage(brokers []string, topic string, message []byte) error {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Printf("Kafka producer creation failed: %v", err)
		return err
	}
	defer producer.Close()

	kafkaMsg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}

	_, _, err = producer.SendMessage(kafkaMsg)
	if err != nil {
		log.Printf("Failed to send message to Kafka: %v", err)
		return err
	}

	log.Printf("Message sent to topic %s", topic)

	return nil
}

