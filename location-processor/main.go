package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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
    GSI2PK    string `json:"gsi2pk" dynamodbav:"GSI2PK"`
    GSI3PK    string `json:"gsi3pk" dynamodbav:"GSI3PK"`
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

var (
    ddb               *dynamodb.DynamoDB
    messagesReceived  atomic.Int64
    messagesProcessed atomic.Int64
    messagesFailedTotal atomic.Int64
    ddbWriteAttempts  atomic.Int64
    ddbWriteSuccesses atomic.Int64
    ddbWriteFailures  atomic.Int64
    batchSize         = 25
    batchInterval     = 1 * time.Second
    batchMutex        sync.Mutex
    itemBatches       = make(map[string][]*dynamodb.WriteRequest)
    batchTimers       = make(map[string]*time.Timer)
)

func main() {
    dynamoEndpoint := "http://localhost:8000"

    sess := session.Must(session.NewSession(&aws.Config{
        Endpoint:    aws.String(dynamoEndpoint),
        Region:      aws.String("us-west-2"),
        Credentials: credentials.NewStaticCredentials("dummy", "dummy", ""),
        DisableSSL:  aws.Bool(true),
    }))
    ddb = dynamodb.New(sess)

    if err := ensureTableExists(); err != nil {
        log.Fatalf("Failed to ensure table exists: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigchan := make(chan os.Signal, 1)
        signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
        <-sigchan
        fmt.Println("\nReceived shutdown signal")
        cancel()
    }()

    var wg sync.WaitGroup

    clusters := []struct {
        name    string
        address string
        topic   string
    }{
        {"Mumbai", "localhost:9101", "mumbai-locations"},
        {"Pune", "localhost:9102", "pune-locations"},
        {"Delhi", "localhost:9103", "delhi-locations"},
    }

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

    go func() {
        ticker := time.NewTicker(30 * time.Second)
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                received := messagesReceived.Load()
                processed := messagesProcessed.Load()
                failed := messagesFailedTotal.Load()
                attempts := ddbWriteAttempts.Load()
                successes := ddbWriteSuccesses.Load()
                failures := ddbWriteFailures.Load()
                
                log.Printf("METRICS REPORT - Kafka: [Received: %d, Processed: %d, Failed: %d] - DynamoDB: [Attempts: %d, Successes: %d, Failures: %d]",
                    received, processed, failed, attempts, successes, failures)
                    
                if received > 0 && processed < received {
                    log.Printf("WARNING: Potential data loss - Only processed %d of %d messages (%.2f%%)",
                        processed, received, float64(processed)/float64(received)*100)
                }
                
                if attempts > 0 && successes < attempts {
                    log.Printf("WARNING: DynamoDB write issues - Success rate: %.2f%%",
                        float64(successes)/float64(attempts)*100)
                }
            case <-ctx.Done():
                return
            }
        }
    }()

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

            messagesReceived.Add(1)
            
            if err := processMessage(msg); err == nil {
                consumer.CommitMessage(msg)
                messagesProcessed.Add(1)
                fmt.Printf("[%s] Processed message from %s (partition %d, offset %d)\n",
                    clusterName, topic,
                    msg.TopicPartition.Partition,
                    msg.TopicPartition.Offset)
            } else {
                messagesFailedTotal.Add(1)
                log.Printf("[%s] Error processing message from %s (partition %d, offset %d): %v", 
                    clusterName, topic, 
                    msg.TopicPartition.Partition, 
                    msg.TopicPartition.Offset, 
                    err)
                
                consumer.CommitMessage(msg)
            }
        }
    }
}

func ensureTableExists() error {
    _, err := ddb.DescribeTable(&dynamodb.DescribeTableInput{
        TableName: aws.String(tableName),
    })
    
    if err != nil {
        if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
            log.Printf("Table %s does not exist, creating it now...", tableName)
            return createTable()
        }
        return fmt.Errorf("error checking table existence: %w", err)
    }
    
    log.Printf("Table %s already exists", tableName)
    return nil
}

func createTable() error {
    input := &dynamodb.CreateTableInput{
        TableName: aws.String(tableName),
        AttributeDefinitions: []*dynamodb.AttributeDefinition{
            {AttributeName: aws.String("PK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("SK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI1PK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI1SK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI2PK"), AttributeType: aws.String("S")},
            {AttributeName: aws.String("GSI3PK"), AttributeType: aws.String("S")},
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
            {
                IndexName: aws.String("StatusH3Res8Index"),
                KeySchema: []*dynamodb.KeySchemaElement{
                    {AttributeName: aws.String("GSI2PK"), KeyType: aws.String("HASH")},
                },
                Projection: &dynamodb.Projection{
                    ProjectionType: aws.String("ALL"),
                },
                ProvisionedThroughput: &dynamodb.ProvisionedThroughput{
                    ReadCapacityUnits:  aws.Int64(10),
                    WriteCapacityUnits: aws.Int64(10),
                },
            },
            {
                IndexName: aws.String("StatusH3Res7Index"),
                KeySchema: []*dynamodb.KeySchemaElement{
                    {AttributeName: aws.String("GSI3PK"), KeyType: aws.String("HASH")},
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
    
    _, err := ddb.CreateTable(input)
    if err != nil {
        return fmt.Errorf("failed to create table: %w", err)
    }
    
    return waitForTableCreation(tableName)
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

    if err := json.Unmarshal(msg.Value, &loc); err != nil {
        log.Printf("Error parsing message: %v. Raw message: %s\n", err, string(msg.Value))
        return fmt.Errorf("failed to unmarshal location: %w", err)
    }

    if loc.DriverID == "" || loc.Latitude < -90 || loc.Latitude > 90 || 
       loc.Longitude < -180 || loc.Longitude > 180 || loc.Status == "" || 
       loc.VehicleType == "" {
        log.Printf("Invalid location data: %+v\n", loc)
        return fmt.Errorf("missing or invalid fields in location data")
    }

    h3Indexes, err := generateH3Indexes(loc.Latitude, loc.Longitude)
    if err != nil {
        log.Printf("Failed to generate H3 indexes for driver %s: %v\n", loc.DriverID, err)
        return err
    }   

    h3Prefix := safePrefix(h3Indexes[0], 5)
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

    var shard string
    if len(loc.DriverID) >= 3 {
        shard = loc.DriverID[len(loc.DriverID)-3:]
    } else {
        shard = loc.DriverID + strings.Repeat("0", 3-len(loc.DriverID))
    }
    driverLoc.PK = fmt.Sprintf("H3#9#%s_%s", h3Prefix, shard)
    driverLoc.SK = fmt.Sprintf("DRIVER#%s#%s", loc.DriverID, loc.Status)
    driverLoc.GSI1PK = fmt.Sprintf("%s#H3#9#%s", loc.Status, h3Indexes[0][:5])
    driverLoc.GSI1SK = fmt.Sprintf("TS#%d", driverLoc.UpdatedAt)
    driverLoc.GSI2PK = fmt.Sprintf("%s#H3#8#%s", loc.Status, h3Indexes[1][:5])
    driverLoc.GSI3PK = fmt.Sprintf("%s#H3#7#%s", loc.Status, h3Indexes[2][:5])

    log.Printf("Generated keys for DriverID %s: PK=%s, SK=%s\n", loc.DriverID, driverLoc.PK, driverLoc.SK)

    return addToBatch(driverLoc)
}

func generateH3Indexes(lat, lon float64) ([]string, error) {
    resolutions := []int{9, 8, 7}
    indexes := make([]string, len(resolutions))

    for i, res := range resolutions {
        index, err := h3.LatLngToCell(h3.LatLng{
            Lat: lat,
            Lng: lon,
        }, res)
        if err != nil {
            return nil, fmt.Errorf("failed to generate H3 index at resolution %d: %w", res, err)
        }
        indexes[i] = strconv.FormatUint(uint64(index), 10)
    }
    return indexes, nil
}

// Helper function for safe prefix extraction
func safePrefix(s string, prefixLen int) string {
    if len(s) >= prefixLen {
        return s[:prefixLen]
    }
    return s + strings.Repeat("0", prefixLen-len(s))
}

func addToBatch(loc DriverLocation) error {
    // Marshal item
    item, err := dynamodbattribute.MarshalMap(loc)
    if err != nil {
        return fmt.Errorf("failed to marshal location: %w", err)
    }
    
    // Create write request
    request := &dynamodb.WriteRequest{
        PutRequest: &dynamodb.PutRequest{
            Item: item,
        },
    }
    
    batchMutex.Lock()
    defer batchMutex.Unlock()
    
    // Add to batch
    itemBatches[tableName] = append(itemBatches[tableName], request)
    
    // If this is the first item, start a timer to flush the batch
    if len(itemBatches[tableName]) == 1 {
        if timer, exists := batchTimers[tableName]; exists {
            timer.Stop()
        }
        batchTimers[tableName] = time.AfterFunc(batchInterval, func() {
            flushBatch(tableName)
        })
    }
    
    // If we've reached batch size, flush immediately
    if len(itemBatches[tableName]) >= batchSize {
        go flushBatch(tableName)
    }
    
    return nil
}

func flushBatch(tableName string) {
    batchMutex.Lock()
    if len(itemBatches[tableName]) == 0 {
        batchMutex.Unlock()
        return
    }
    
    // Get the batch and clear it while holding the lock
    batch := itemBatches[tableName]
    itemBatches[tableName] = nil
    if timer, exists := batchTimers[tableName]; exists {
        timer.Stop()
        delete(batchTimers, tableName)
    }
    batchMutex.Unlock()
    
    // Now process the batch without holding the lock
    log.Printf("Flushing batch of %d items to DynamoDB", len(batch))
    
    // Split into chunks of 25 items (DynamoDB batch limit)
    for i := 0; i < len(batch); i += 25 {
        end := i + 25
        if end > len(batch) {
            end = len(batch)
        }
        
        chunk := batch[i:end]
        requestItems := map[string][]*dynamodb.WriteRequest{
            tableName: chunk,
        }
        
        // Use BatchWriteItem with retry logic
        maxRetries := 10
        for attempt := 1; attempt <= maxRetries; attempt++ {
            ddbWriteAttempts.Add(1)
            out, err := ddb.BatchWriteItem(&dynamodb.BatchWriteItemInput{
                RequestItems: requestItems,
            })
            
            if err == nil && (out.UnprocessedItems == nil || len(out.UnprocessedItems) == 0) {
                // Success with all items
                ddbWriteSuccesses.Add(int64(len(chunk)))
                break
            }
            
            if err != nil {
                log.Printf("Batch write error (attempt %d/%d): %v", attempt, maxRetries, err)
                
                if attempt == maxRetries {
                    log.Printf("Failed to write batch after %d attempts", maxRetries)
                    ddbWriteFailures.Add(int64(len(chunk)))
                    break
                }
                
                // Wait before retry
                time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond)
                continue
            }
            
            // Handle unprocessed items
            if len(out.UnprocessedItems) > 0 {
                processedCount := len(chunk) - len(out.UnprocessedItems[tableName])
                ddbWriteSuccesses.Add(int64(processedCount))
                
                log.Printf("%d items were unprocessed in batch, retrying", len(out.UnprocessedItems[tableName]))
                requestItems = out.UnprocessedItems
                
                if attempt == maxRetries {
                    log.Printf("Failed to process all items after %d attempts, %d items remaining", 
                        maxRetries, len(out.UnprocessedItems[tableName]))
                    ddbWriteFailures.Add(int64(len(out.UnprocessedItems[tableName])))
                    break
                }
                
                time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond)
                continue
            }
        }
    }
}
