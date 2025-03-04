package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

const (
	KafkaBroker   = "localhost:9092"
	DriverTopic   = "driver-location-updates"
	DriverConsumerGroup = "driver-consumer-group"
)

type Consumer struct{}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (c *Consumer) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Println("Received message:", string(msg.Value))
		sess.MarkMessage(msg, "") 
	}
	return nil
}

func main() {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.AutoCommit.Enable = true

	consumerGroup, err := sarama.NewConsumerGroup([]string{KafkaBroker}, DriverConsumerGroup, config)
	if err != nil {
		panic(err)
	}
	defer consumerGroup.Close()

	fmt.Println("Consumer started")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		cancel()
	}()

	for {
		err := consumerGroup.Consume(ctx, []string{DriverTopic}, &Consumer{})
		if err != nil {
			fmt.Println("Error from consumer:", err)
		}

		if ctx.Err() != nil {
			break
		}
	}
}
