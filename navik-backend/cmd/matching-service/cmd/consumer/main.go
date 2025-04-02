package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matching-service/internal/config"
	"matching-service/internal/handler"
	"matching-service/internal/repository"
	"matching-service/internal/service"

	"github.com/IBM/sarama"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint: aws.String(cfg.DynamoDB.Endpoint),
		Region:   aws.String(cfg.DynamoDB.Region),
		Credentials: credentials.NewStaticCredentials(
			cfg.DynamoDB.AccessKey,
			cfg.DynamoDB.SecretKey,
			""),
		DisableSSL: aws.Bool(true),
	}))
	ddb := dynamodb.New(sess)

	// Create driver repository
	driverRepo := repository.NewDriverRepository(ddb, cfg.DynamoDB.TableName)

	// Create matching service
	matchingService := service.NewMatchingService(driverRepo, struct {
		MinDriversToReturn int
		MaxDistanceKm      float64
	}{
		MinDriversToReturn: cfg.Matching.MinDriversToReturn,
		MaxDistanceKm:      cfg.Matching.MaxDistanceKm,
	})

	// Setup Kafka consumer config
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Version = sarama.V2_8_0_0
	
	// Set to read from the oldest message when no committed offset exists
	kafkaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	
	// Increase max wait time to batch more messages (optional performance tuning)
	kafkaConfig.Consumer.MaxWaitTime = 500 * time.Millisecond

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.GroupID, kafkaConfig)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}
	defer consumerGroup.Close()
	
	// Create error handling channel
	consumerErrors := make(chan error, 1)
	go func() {
		for err := range consumerGroup.Errors() {
			log.Printf("Consumer group error: %v", err)
		}
	}()

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := handler.NewConsumerGroupHandler(matchingService)

	// Start consuming in a goroutine
	go func() {
		for {
			log.Printf("Starting consumer for topics: %v", cfg.Kafka.Topics)
			
			if err := consumerGroup.Consume(ctx, cfg.Kafka.Topics, handler); err != nil {
				if err == sarama.ErrClosedConsumerGroup {
					log.Println("Consumer group has been closed")
					return
				}
				
				log.Printf("Error from consumer: %v", err)
				select {
				case consumerErrors <- err:
				default:
				}
				
				time.Sleep(5 * time.Second)
				continue
			}

			if ctx.Err() != nil {
				return
			}
			
			log.Println("Consumer group session ended, rebalancing")
		}
	}()

	log.Println("Matching service started. Consuming messages from the beginning of topics...")

	<-signals
	log.Println("Received termination signal. Shutting down...")
	cancel()
	
	// Wait for all in-flight messages to be processed
	time.Sleep(2 * time.Second)
}
