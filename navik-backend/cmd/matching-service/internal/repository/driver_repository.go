package repository

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"matching-service/internal/model"
	"matching-service/internal/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// DriverRepository defines the interface for driver data access
type DriverRepository interface {
	FindDriversInH9Cell(ctx context.Context, h3Index string) ([]model.DriverLocation, error)
	FindDriversInH9Cells(ctx context.Context, h3Indices []string) ([]model.DriverLocation, error)
	FindDriversInH8Cell(ctx context.Context, h3Index string) ([]model.DriverLocation, error)
	FindDriversInH8Cells(ctx context.Context, h3Indices []string) ([]model.DriverLocation, error)
	FindDriversInH7Cell(ctx context.Context, h3Index string) ([]model.DriverLocation, error)
	FindDriversInH7Cells(ctx context.Context, h3Indices []string) ([]model.DriverLocation, error)
}

type driverRepository struct {
	ddb       *dynamodb.DynamoDB
	tableName string
}

// NewDriverRepository creates a new driver repository
func NewDriverRepository(ddb *dynamodb.DynamoDB, tableName string) DriverRepository {
	return &driverRepository{
		ddb:       ddb,
		tableName: tableName,
	}
}

// FindDriversInH9Cell queries the database for drivers in a specific H9 cell
func (r *driverRepository) FindDriversInH9Cell(ctx context.Context, h3Index string) ([]model.DriverLocation, error) {
	h3Prefix := util.SafePrefix(h3Index, 5)
	gsi1pk := fmt.Sprintf("ACTIVE#H3#9#%s", h3Prefix)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("StatusH3Index"),
		KeyConditionExpression: aws.String("GSI1PK = :gsi1pk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi1pk": {
				S: aws.String(gsi1pk),
			},
		},
	}

	result, err := r.ddb.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	return r.parseDriverItems(result.Items)
}

// FindDriversInH9Cells queries DynamoDB for drivers in multiple H9 cells
func (r *driverRepository) FindDriversInH9Cells(ctx context.Context, h3Indices []string) ([]model.DriverLocation, error) {
	if len(h3Indices) == 0 {
		return []model.DriverLocation{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allDrivers []model.DriverLocation
	var queryErrors []error

	for _, h3Index := range h3Indices {
		wg.Add(1)
		go func(index string) {
			defer wg.Done()

			drivers, err := r.FindDriversInH9Cell(ctx, index)
			if err != nil {
				mu.Lock()
				queryErrors = append(queryErrors, fmt.Errorf("failed to query H9 cell %s: %w", index, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			allDrivers = append(allDrivers, drivers...)
			mu.Unlock()
		}(h3Index)
	}

	wg.Wait()

	if len(queryErrors) > 0 {
		return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
	}

	return allDrivers, nil
}

// FindDriversInH8Cell queries DynamoDB for drivers in a specific H8 cell
func (r *driverRepository) FindDriversInH8Cell(ctx context.Context, h8Index string) ([]model.DriverLocation, error) {
	h8Prefix := util.SafePrefix(h8Index, 5)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("StatusH3Res8Index"),
		KeyConditionExpression: aws.String("GSI2PK = :gsi2pk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi2pk": {
				S: aws.String(fmt.Sprintf("ACTIVE#H3#8#%s", h8Prefix)),
			},
		},
	}

	result, err := r.ddb.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB for H8 cell: %w", err)
	}

	return r.parseDriverItems(result.Items)
}

// FindDriversInH8Cells queries DynamoDB for drivers in multiple H8 cells
func (r *driverRepository) FindDriversInH8Cells(ctx context.Context, h8Indices []string) ([]model.DriverLocation, error) {
	if len(h8Indices) == 0 {
		return []model.DriverLocation{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allDrivers []model.DriverLocation
	var queryErrors []error

	for _, h8Index := range h8Indices {
		wg.Add(1)
		go func(index string) {
			defer wg.Done()

			drivers, err := r.FindDriversInH8Cell(ctx, index)
			if err != nil {
				mu.Lock()
				queryErrors = append(queryErrors, fmt.Errorf("failed to query H8 cell %s: %w", index, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			allDrivers = append(allDrivers, drivers...)
			mu.Unlock()
		}(h8Index)
	}

	wg.Wait()

	if len(queryErrors) > 0 {
		return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
	}

	return allDrivers, nil
}

// FindDriversInH7Cell queries DynamoDB for drivers in a specific H7 cell
func (r *driverRepository) FindDriversInH7Cell(ctx context.Context, h7Index string) ([]model.DriverLocation, error) {
	h7Prefix := util.SafePrefix(h7Index, 5)

	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.tableName),
		IndexName:              aws.String("StatusH3Res7Index"),
		KeyConditionExpression: aws.String("GSI3PK = :gsi3pk"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gsi3pk": {
				S: aws.String(fmt.Sprintf("ACTIVE#H3#7#%s", h7Prefix)),
			},
		},
	}

	result, err := r.ddb.QueryWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query DynamoDB for H7 cell: %w", err)
	}

	return r.parseDriverItems(result.Items)
}

// FindDriversInH7Cells queries DynamoDB for drivers in multiple H7 cells
func (r *driverRepository) FindDriversInH7Cells(ctx context.Context, h7Indices []string) ([]model.DriverLocation, error) {
	if len(h7Indices) == 0 {
		return []model.DriverLocation{}, nil
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var allDrivers []model.DriverLocation
	var queryErrors []error

	for _, h7Index := range h7Indices {
		wg.Add(1)
		go func(index string) {
			defer wg.Done()

			drivers, err := r.FindDriversInH7Cell(ctx, index)
			if err != nil {
				mu.Lock()
				queryErrors = append(queryErrors, fmt.Errorf("failed to query H7 cell %s: %w", index, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			allDrivers = append(allDrivers, drivers...)
			mu.Unlock()
		}(h7Index)
	}

	wg.Wait()

	if len(queryErrors) > 0 {
		return allDrivers, fmt.Errorf("errors occurred during queries: %v", queryErrors)
	}

	return allDrivers, nil
}

// Helper function to parse DynamoDB items into driver locations
func (r *driverRepository) parseDriverItems(items []map[string]*dynamodb.AttributeValue) ([]model.DriverLocation, error) {
	var drivers []model.DriverLocation

	for _, item := range items {
		if item["location"] == nil || item["location"].S == nil {
			log.Printf("Warning: Item missing location field")
			continue
		}

		locString := *item["location"].S
		locParts := strings.Split(locString, ",")

		if len(locParts) != 2 {
			log.Printf("Warning: Invalid location format: %s", locString)
			continue
		}

		lat, err := strconv.ParseFloat(locParts[0], 64)
		if err != nil {
			log.Printf("Warning: Failed to parse latitude: %s", locParts[0])
			continue
		}

		lng, err := strconv.ParseFloat(locParts[1], 64)
		if err != nil {
			log.Printf("Warning: Failed to parse longitude: %s", locParts[1])
			continue
		}

		var driver model.DriverLocation
		err = dynamodbattribute.UnmarshalMap(item, &driver)
		if err != nil {
			log.Printf("Warning: Failed to unmarshal driver: %v", err)
			continue
		}

		driver.Latitude = lat
		driver.Longitude = lng

		drivers = append(drivers, driver)
	}

	return drivers, nil
}
