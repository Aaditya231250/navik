package repository

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/uber/h3-go/v4"

	"location-service/internal/model"
)

const (
	tableName     = "driver-locations"
	locationTTL   = 900
	batchSize     = 25
	batchInterval = 1 * time.Second
	maxRetries    = 10
)

type DynamoDBLocationRepository struct {
	ddb            *dynamodb.DynamoDB
	batchMutex     sync.Mutex
	itemBatches    map[string][]*dynamodb.WriteRequest
	batchTimers    map[string]*time.Timer
	writeAttempts  *int64
	writeSuccesses *int64
	writeFailures  *int64
}

func NewDynamoDBLocationRepository(ddb *dynamodb.DynamoDB, writeAttempts, writeSuccesses, writeFailures *int64) *DynamoDBLocationRepository {
	return &DynamoDBLocationRepository{
		ddb:            ddb,
		itemBatches:    make(map[string][]*dynamodb.WriteRequest),
		batchTimers:    make(map[string]*time.Timer),
		writeAttempts:  writeAttempts,
		writeSuccesses: writeSuccesses,
		writeFailures:  writeFailures,
	}
}

func (r *DynamoDBLocationRepository) Store(ctx context.Context, loc model.Location) error {
	locDB, err := r.convertToLocationDB(loc)
	if err != nil {
		return fmt.Errorf("failed to convert location: %w", err)
	}

	return r.addToBatch(locDB)
}

// convertToLocationDB transforms Location to LocationDB with H3 indexing
func (r *DynamoDBLocationRepository) convertToLocationDB(loc model.Location) (model.LocationDB, error) {
	h3Indexes, err := r.generateH3Indexes(loc.Latitude, loc.Longitude)
	if err != nil {
		return model.LocationDB{}, fmt.Errorf("failed to generate H3 indexes: %w", err)
	}

	h3Prefix := r.safePrefix(h3Indexes[0], 5)

	locDB := model.LocationDB{
		DriverID:    loc.DriverID,
		Location:    fmt.Sprintf("%f,%f", loc.Latitude, loc.Longitude),
		H3Res9:      h3Indexes[0],
		H3Res8:      h3Indexes[1],
		H3Res7:      h3Indexes[2],
		VehicleType: loc.VehicleType,
		Status:      loc.Status,
		UpdatedAt:   time.Now().Unix(),
		ExpiresAt:   time.Now().Unix() + locationTTL,
	}

	var shard string
	if len(loc.DriverID) >= 3 {
		shard = loc.DriverID[len(loc.DriverID)-3:]
	} else {
		shard = loc.DriverID + strings.Repeat("0", 3-len(loc.DriverID))
	}

	locDB.PK = fmt.Sprintf("H3#9#%s_%s", h3Prefix, shard)
	locDB.SK = fmt.Sprintf("DRIVER#%s#%s", loc.DriverID, loc.Status)
	locDB.GSI1PK = fmt.Sprintf("%s#H3#9#%s", loc.Status, h3Indexes[0][:5])
	locDB.GSI1SK = fmt.Sprintf("TS#%d", locDB.UpdatedAt)
	locDB.GSI2PK = fmt.Sprintf("%s#H3#8#%s", loc.Status, h3Indexes[1][:5])
	locDB.GSI3PK = fmt.Sprintf("%s#H3#7#%s", loc.Status, h3Indexes[2][:5])

	log.Printf("Generated keys for DriverID %s: PK=%s, SK=%s\n", locDB.DriverID, locDB.PK, locDB.SK)

	return locDB, nil
}

func (r *DynamoDBLocationRepository) generateH3Indexes(lat, lon float64) ([]string, error) {
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

func (r *DynamoDBLocationRepository) safePrefix(s string, prefixLen int) string {
	if len(s) >= prefixLen {
		return s[:prefixLen]
	}
	return s + strings.Repeat("0", prefixLen-len(s))
}

func (r *DynamoDBLocationRepository) addToBatch(loc model.LocationDB) error {
	item, err := dynamodbattribute.MarshalMap(loc)
	if err != nil {
		return fmt.Errorf("failed to marshal location: %w", err)
	}

	request := &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: item,
		},
	}

	r.batchMutex.Lock()
	defer r.batchMutex.Unlock()

	r.itemBatches[tableName] = append(r.itemBatches[tableName], request)

	// Start timer if first item in batch
	if len(r.itemBatches[tableName]) == 1 {
		if timer, exists := r.batchTimers[tableName]; exists {
			timer.Stop()
		}
		r.batchTimers[tableName] = time.AfterFunc(batchInterval, func() {
			r.flushBatch(tableName)
		})
	}

	// Flush immediately if batch is full
	if len(r.itemBatches[tableName]) >= batchSize {
		go r.flushBatch(tableName)
	}

	return nil
}

func (r *DynamoDBLocationRepository) flushBatch(tableName string) {
	r.batchMutex.Lock()
	if len(r.itemBatches[tableName]) == 0 {
		r.batchMutex.Unlock()
		return
	}

	// Get the batch and clear it while holding the lock
	batch := r.itemBatches[tableName]
	r.itemBatches[tableName] = nil
	if timer, exists := r.batchTimers[tableName]; exists {
		timer.Stop()
		delete(r.batchTimers, tableName)
	}
	r.batchMutex.Unlock()

	log.Printf("Flushing batch of %d items to DynamoDB", len(batch))

	// Process batch in chunks of 25 (DynamoDB limit)
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
		r.writeBatchWithRetry(requestItems, chunk)
	}
}

func (r *DynamoDBLocationRepository) writeBatchWithRetry(requestItems map[string][]*dynamodb.WriteRequest, chunk []*dynamodb.WriteRequest) {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		atomic.AddInt64(r.writeAttempts, 1)

		out, err := r.ddb.BatchWriteItem(&dynamodb.BatchWriteItemInput{
			RequestItems: requestItems,
		})

		if err == nil && (out.UnprocessedItems == nil || len(out.UnprocessedItems) == 0) {
			// All items processed successfully
			atomic.AddInt64(r.writeSuccesses, int64(len(chunk)))
			break
		}

		if err != nil {
			log.Printf("Batch write error (attempt %d/%d): %v", attempt, maxRetries, err)

			if attempt == maxRetries {
				log.Printf("Failed to write batch after %d attempts", maxRetries)
				atomic.AddInt64(r.writeFailures, int64(len(chunk)))
				break
			}

			// Exponential backoff
			time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond)
			continue
		}

		// Handle unprocessed items
		if len(out.UnprocessedItems) > 0 {
			processedCount := len(chunk) - len(out.UnprocessedItems[tableName])
			atomic.AddInt64(r.writeSuccesses, int64(processedCount))

			log.Printf("%d items were unprocessed in batch, retrying", len(out.UnprocessedItems[tableName]))
			requestItems = out.UnprocessedItems

			if attempt == maxRetries {
				log.Printf("Failed to process all items after %d attempts, %d items remaining",
					maxRetries, len(out.UnprocessedItems[tableName]))
				atomic.AddInt64(r.writeFailures, int64(len(out.UnprocessedItems[tableName])))
				break
			}

			time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * 200 * time.Millisecond)
		}
	}
}

// EnsureTableExists creates the DynamoDB table if it doesn't exist
func (r *DynamoDBLocationRepository) EnsureTableExists() error {
	_, err := r.ddb.DescribeTable(&dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == dynamodb.ErrCodeResourceNotFoundException {
			log.Printf("Table %s does not exist, creating it now...", tableName)
			return r.createTable()
		}
		return fmt.Errorf("error checking table existence: %w", err)
	}

	log.Printf("Table %s already exists", tableName)
	return nil
}

func (r *DynamoDBLocationRepository) createTable() error {
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

	_, err := r.ddb.CreateTable(input)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return r.waitForTableCreation(tableName)
}

func (r *DynamoDBLocationRepository) waitForTableCreation(tableName string) error {
	for i := 0; i < 30; i++ {
		resp, err := r.ddb.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err == nil && *resp.Table.TableStatus == "ACTIVE" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("table did not become active in time")
}
