package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func CreateUsersTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	log.Printf("Checking if table %s exists...", tableName)
	
	// Try to describe the table first to see if it exists
	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	})
	
	// If the error is nil, table exists
	if err == nil {
		log.Printf("Table %s already exists", tableName)
		return nil
	}
	
	// Check if the error is because the table doesn't exist
	var notFoundErr *types.ResourceNotFoundException
	if !errors.As(err, &notFoundErr) {
		// Some other error occurred
		return fmt.Errorf("failed to check if table exists: %w", err)
	}
	
	// Table doesn't exist, create it
	log.Printf("Creating table %s...", tableName)
	_, err = client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("email"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("id"),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("EmailIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("email"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
	})

	if err != nil {
		var resourceInUseErr *types.ResourceInUseException
		if errors.As(err, &resourceInUseErr) {
			log.Printf("Table %s was created by another process", tableName)
			return nil
		}
		return fmt.Errorf("failed to create table: %w", err)
	}

	log.Printf("Created table %s successfully", tableName)
	return nil
}
