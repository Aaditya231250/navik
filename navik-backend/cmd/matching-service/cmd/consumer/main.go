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

	// Create consumer group
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.GroupID, kafkaConfig)
	if err != nil {
		log.Fatalf("Error creating consumer group: %v", err)
	}
	defer consumerGroup.Close()

	// Setup signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	handler := handler.NewConsumerGroupHandler(matchingService)

	// Start consuming in a goroutine
	go func() {
		for {
			if err := consumerGroup.Consume(ctx, cfg.Kafka.Topics, handler); err != nil {
				log.Printf("Error from consumer: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("Matching service started. Waiting for user requests...")

	<-signals
	log.Println("Received termination signal. Shutting down...")
	cancel()
}
