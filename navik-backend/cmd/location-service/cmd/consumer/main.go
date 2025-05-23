package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"

	"location-service/internal/model"
	"location-service/internal/repository"
	"location-service/internal/service"
	"location-service/pkg/kafka"
)

func main() {
	dynamoEndpoint := flag.String("dynamo-endpoint", "http://dynamodb-local:8000", "DynamoDB endpoint")
	flag.Parse()

	var messagesReceived, messagesProcessed, messagesFailedTotal int64
	var ddbWriteAttempts, ddbWriteSuccesses, ddbWriteFailures int64

	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(*dynamoEndpoint),
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials("localkey", "localsecret", ""),
		DisableSSL:  aws.Bool(true),
	}))
	ddb := dynamodb.New(sess)
	log.Println("DynamoDB session initialized")

	_, err := ddb.ListTables(&dynamodb.ListTablesInput{})
	if err != nil {
		log.Printf("ERROR: Failed to connect to DynamoDB: %v", err)
		log.Printf("Check if DynamoDB is running and accessible at %s", *dynamoEndpoint)
	} else {
		log.Println("Successfully connected to DynamoDB")
	}

	repo := repository.NewDynamoDBLocationRepository(ddb, &ddbWriteAttempts, &ddbWriteSuccesses, &ddbWriteFailures)

	if err := repo.EnsureTableExists(); err != nil {
		log.Fatalf("Failed to ensure table exists: %v", err)
	}
	log.Println("DynamoDB table is ready")

	locationService := service.NewLocationService(repo, nil)

	// Handler function for location updates
	locationHandler := func(loc model.Location) error {
		return locationService.ProcessLocationUpdate(loc)
	}

	clusters := []kafka.ClusterConfig{
		{Name: "Mumbai", BootstrapServers: "kafka-mumbai:29092", Topic: "mumbai-locations"},
		{Name: "Pune", BootstrapServers: "kafka-pune:29092", Topic: "pune-locations"},
		{Name: "Delhi", BootstrapServers: "kafka-delhi:29092", Topic: "delhi-locations"},
	}

	consumer, err := kafka.NewConsumer("driver-location-consumer", clusters, locationHandler,
		&messagesReceived, &messagesProcessed, &messagesFailedTotal)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}
	log.Println("Kafka consumer is ready")
	defer consumer.Close()

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				reportMetrics(&messagesReceived, &messagesProcessed, &messagesFailedTotal,
					&ddbWriteAttempts, &ddbWriteSuccesses, &ddbWriteFailures)
			}
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start consumer in background
	go func() {
		if err := consumer.Consume(ctx); err != nil {
			log.Fatalf("Error in consumer: %v", err)
		}
	}()

	// Wait for shutdown signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down...")
	cancel() // Stop the consumer
	log.Println("Service stopped gracefully")
}

func reportMetrics(received, processed, failed, attempts, successes, failures *int64) {
	r := atomic.LoadInt64(received)
	p := atomic.LoadInt64(processed)
	f := atomic.LoadInt64(failed)
	a := atomic.LoadInt64(attempts)
	s := atomic.LoadInt64(successes)
	fa := atomic.LoadInt64(failures)

	log.Printf("METRICS REPORT - Kafka: [Received: %d, Processed: %d, Failed: %d] - DynamoDB: [Attempts: %d, Successes: %d, Failures: %d]",
		r, p, f, a, s, fa)

	if r > 0 && p < r {
		log.Printf("WARNING: Potential data loss - Only processed %d of %d messages (%.2f%%)",
			p, r, float64(p)/float64(r)*100)
	}

	if a > 0 && s < a {
		log.Printf("WARNING: DynamoDB write issues - Success rate: %.2f%%",
			float64(s)/float64(a)*100)
	}
}
