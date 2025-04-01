package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/uber/h3-go/v4"
)

const (
	tableName     = "DriverLocations"
	locationTTL   = 900 // 15 minutes in seconds
	consumerGroup = "driver-location-consumer"
)

type DriverLocation struct {
    PK        string `json:"pk" dynamodbav:"PK"`
    SK        string `json:"sk" dynamodbav:"SK"`
    GSI1PK    string `json:"gsi1pk" dynamodbav:"GSI1PK"`
    GSI1SK    string `json:"gsi1sk" dynamodbav:"GSI1SK"`
    DriverID  string `json:"driver_id" dynamodbav:"driver_id"`
    Location  string `json:"location" dynamodbav:"location"`
    H3Res9    string `json:"h3_res9" dynamodbav:"h3_res9"`
    H3Res8    string `json:"h3_res8" dynamodbav:"h3_res8"`
    H3Res7    string `json:"h3_res7" dynamodbav:"h3_res7"`
    Vehicle   string `json:"vehicle_type" dynamodbav:"vehicle_type"`
    Status    string `json:"status" dynamodbav:"status"`
    UpdatedAt int64  `json:"updated_at" dynamodbav:"updated_at"`
    ExpiresAt int64  `json:"expires_at" dynamodbav:"expires_at"`
}

type Location struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var ddb *dynamodb.DynamoDB

// ddb               *dynamodb.DynamoDB
// messagesReceived  atomic.Int64
// messagesProcessed atomic.Int64
// messagesFailedTotal atomic.Int64
// ddbWriteAttempts  atomic.Int64
// ddbWriteSuccesses atomic.Int64
// ddbWriteFailures  atomic.Int64
// batchSize         = 25
// batchInterval     = 1 * time.Second
// batchMutex        sync.Mutex
// itemBatches       = make(map[string][]*dynamodb.WriteRequest)
// batchTimers       = make(map[string]*time.Timer)

func main() {
	dynamoEndpoint := "http://localhost:8000"

	// Initialize DynamoDB
	sess := session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(dynamoEndpoint),
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
		DisableSSL:  aws.Bool(true),
	}))
	ddb = dynamodb.New(sess)

	// Create table if not exists
	if err := recreateTable(); err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Wait for table to become active
	if err := waitForTableCreation(tableName); err != nil {
		log.Fatalf("Table did not become active: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
		<-sigchan
		fmt.Println("\nReceived shutdown signal")
		cancel()
	}()

	var wg sync.WaitGroup

	// Configure Kafka clusters
	clusters := []struct {
		name    string
		address string
		topic   string
	}{
		{"Mumbai", "localhost:9101", "mumbai-locations"},
		{"Pune", "localhost:9102", "pune-locations"},
		{"Delhi", "localhost:9103", "delhi-locations"},
	}

	// Start a consumer for each cluster
	for _, cluster := range clusters {
		wg.Add(1)
		go func(c struct {
			name    string
			address string
			topic   string
		}) {
			defer wg.Done()
			startConsumer(ctx, c.name, c.address, c.topic)
		}(cluster)
	}

	wg.Wait()
	fmt.Println("All consumers stopped")
}

func startConsumer(ctx context.Context, clusterName, bootstrapServers, topic string) {
	config := &kafka.ConfigMap{
		"bootstrap.servers":  bootstrapServers,
		"group.id":           consumerGroup,
		"auto.offset.reset":  "earliest",
		"enable.auto.commit": "false",
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		log.Printf("[%s] Failed to create consumer: %v", clusterName, err)
		return
	}
	defer consumer.Close()

	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		log.Printf("[%s] Failed to subscribe to topic: %v", clusterName, err)
		return
	}

	log.Printf("[%s] Started consumer for topic %s", clusterName, topic)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[%s] Shutting down consumer", clusterName)
			return
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() != kafka.ErrTimedOut {
					log.Printf("[%s] Consumer error: %v", clusterName, err)
				}
				continue
			}

			if err := processMessage(msg); err == nil {
				consumer.CommitMessage(msg)
				fmt.Printf("[%s] Processed message from %s (partition %d, offset %d)\n",
					clusterName, topic,
					msg.TopicPartition.Partition,
					msg.TopicPartition.Offset)
			} else {
				log.Printf("[%s] Error processing message: %v", clusterName, err)
			}
		}
	}
}

func waitForTableCreation(tableName string) error {
	for i := 0; i < 30; i++ {
		resp, err := ddb.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err == nil && *resp.Table.TableStatus == "ACTIVE" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("table did not become active in time")
}

func processMessage(msg *kafka.Message) error {
	var loc struct {
		DriverID    string  `json:"driver_id"`
		City        string  `json:"city"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Timestamp   int64   `json:"timestamp"`
		VehicleType string  `json:"vehicle_type"`
		Status      string  `json:"status"`
	}

	// Unmarshal Kafka message
	if err := json.Unmarshal(msg.Value, &loc); err != nil {
		log.Printf("Error parsing message: %v. Raw message: %s\n", err, string(msg.Value))
		return fmt.Errorf("failed to unmarshal location: %w", err)
	}

	// Validate required fields
	if loc.DriverID == "" || loc.Latitude == 0 || loc.Longitude == 0 || loc.Status == "" || loc.VehicleType == "" {
		log.Printf("Invalid location data: %+v\n", loc)
		return fmt.Errorf("missing required fields in location data")
	}

	// Generate H3 indexes
	h3Indexes := generateH3Indexes(loc.Latitude, loc.Longitude)

	// Create DynamoDB item
	driverLoc := DriverLocation{
		DriverID:  loc.DriverID,
		Location:  fmt.Sprintf("%f,%f", loc.Latitude, loc.Longitude),
		H3Res9:    h3Indexes[0],
		H3Res8:    h3Indexes[1],
		H3Res7:    h3Indexes[2],
		Status:    loc.Status,
		Vehicle:   loc.VehicleType,
		UpdatedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Unix() + locationTTL,
	}

	// Shard partition key
	var shard string
	if len(loc.DriverID) >= 3 {
		shard = loc.DriverID[len(loc.DriverID)-3:]
	} else {
		shard = loc.DriverID + strings.Repeat("0", 3-len(loc.DriverID))
	}
	driverLoc.PK = fmt.Sprintf("H3#9#%s_%s", h3Indexes[0][:5], shard)
	driverLoc.SK = fmt.Sprintf("DRIVER#%s#%s", loc.DriverID, loc.Status)
	driverLoc.GSI1PK = fmt.Sprintf("%s#H3#9#%s", loc.Status, h3Indexes[0][:5])
	driverLoc.GSI1SK = fmt.Sprintf("TS#%d", driverLoc.UpdatedAt)

	// Log generated keys for debugging
	log.Printf("Generated keys for DriverID %s: PK=%s, SK=%s\n", loc.DriverID, driverLoc.PK, driverLoc.SK)

	// Write to DynamoDB
	return writeToDynamoDB(driverLoc)
}

func writeToDynamoDB(loc DriverLocation) error {
    // Validate required keys
    if loc.PK == "" || loc.SK == "" {
        log.Printf("Missing required keys for DynamoDB item: %+v\n", loc)
        return fmt.Errorf("missing required keys for DynamoDB item")
    }

    // Marshal item
    item, err := dynamodbattribute.MarshalMap(loc)
    if err != nil {
        return fmt.Errorf("failed to marshal location: %w", err)
    }

    // Simple put with no conditions
    input := &dynamodb.PutItemInput{
        TableName: aws.String(tableName),
        Item:      item,
    }

    _, err = ddb.PutItem(input)
    if err != nil {
        log.Printf("DynamoDB put failed for item %+v with error %v\n", loc, err)
        return fmt.Errorf("dynamodb put failed: %w", err)
    }
    
    return nil
}

func generateH3Indexes(lat, lon float64) []string {
	resolutions := []int{9, 8, 7}
	indexes := make([]string, len(resolutions))

	for i, res := range resolutions {
		index, err := h3.LatLngToCell(h3.LatLng{
			Lat: lat,
			Lng: lon,
		}, res)
		if err != nil {
			panic(fmt.Sprintf("failed to generate H3 index: %v", err))
		}
		indexes[i] = strconv.FormatUint(uint64(index), 10) // Convert h3.Cell to uint64
	}
	return indexes
}

func isTableExistsError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == dynamodb.ErrCodeResourceInUseException
	}
	return false
}

func recreateTable() error {
    // Delete existing table if it exists
    _, err := ddb.DeleteTable(&dynamodb.DeleteTableInput{
        TableName: aws.String(tableName),
    })
    
    // Ignore error if table doesn't exist
    if err != nil {
        if awsErr, ok := err.(awserr.Error); !ok || awsErr.Code() != dynamodb.ErrCodeResourceNotFoundException {
            log.Printf("Warning during table deletion: %v", err)
        }
    }
    
    // Wait for table deletion to complete
    time.Sleep(5 * time.Second)
    
    // Create new table with correct schema
    input := &dynamodb.CreateTableInput{
        TableName: aws.String(tableName),
        AttributeDefinitions: []*dynamodb.AttributeDefinition{
            {AttributeName: aws.String("PK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("SK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI1PK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI1SK"), AttributeType: aws.String("S")},
        },
        KeySchema: []*dynamodb.KeySchemaElement{
            {AttributeName: aws.String("PK"), KeyType: aws.String("HASH")},
            {AttributeName: aws.String("SK"), KeyType: aws.String("RANGE")},
        },
        GlobalSecondaryIndexes: []*dynamodb.GlobalSecondaryIndex{
            {
                IndexName: aws.String("StatusH3Index"),
                KeySchema: []*dynamodb.KeySchemaElement{
                    {AttributeName: aws.String("GSI1PK"), KeyType: aws.String("HASH")},
                    {AttributeName: aws.String("GSI1SK"), KeyType: aws.String("RANGE")},
                },
                Projection: &dynamodb.Projection{
                    ProjectionType: aws.String("ALL"),
                },
                ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
                    ReadCapacityUnits:  aws.Int64(10),
                    WriteCapacityUnits: aws.Int64(10),
                },
            },
        },
        ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
            ReadCapacityUnits:  aws.Int64(10),
            WriteCapacityUnits: aws.Int64(10),
        },
    }
    
    _, err = ddb.CreateTable(input)
    if err != nil {
        return fmt.Errorf("failed to create table: %w", err)
    }
    
    return waitForTableCreation(tableName)
}
