package repository

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"authentication/internal/domain"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailExists     = errors.New("email already exists")
	ErrInvalidUserType = errors.New("invalid user type")
)

type UserRepository struct {
	client    *dynamodb.Client
	tableName string
}

func NewUserRepository(client *dynamodb.Client, tableName string) *UserRepository {
	return &UserRepository{
		client:    client,
		tableName: tableName,
	}
}

// CreateUser creates a new user in DynamoDB
func (r *UserRepository) CreateUser(ctx context.Context, user domain.User) (string, error) {
	// Check if email already exists
	exists, err := r.emailExists(ctx, user.Email)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrEmailExists
	}

	// Generate ID if not provided
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Marshal user to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return "", err
	}

	// Put item in DynamoDB
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		return "", err
	}

	return user.ID, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	// Create expression for query
	expr, err := expression.NewBuilder().
		WithFilter(expression.Name("email").Equal(expression.Value(email))).
		Build()
	if err != nil {
		return nil, err
	}

	// Scan DynamoDB table with filter
	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, ErrUserNotFound
	}

	// Unmarshal the item to a User struct
	var user domain.User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	// Get item from DynamoDB
	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, ErrUserNotFound
	}

	// Unmarshal the item to a User struct
	var user domain.User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (r *UserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	// Update timestamp
	user.UpdatedAt = time.Now()

	// Marshal user to DynamoDB item
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	// Remove read-only attributes that shouldn't be updated
	delete(item, "created_at")

	// Update expression
	update := expression.Set(expression.Name("updated_at"), expression.Value(user.UpdatedAt))

	// Add all other attributes to update expression
	for k, v := range item {
		if k != "id" && k != "created_at" {
			update = update.Set(expression.Name(k), expression.Value(v))
		}
	}

	expr, err := expression.NewBuilder().
		WithUpdate(update).
		Build()
	if err != nil {
		return err
	}

	// Update item in DynamoDB
	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: user.ID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	return err
}

// UpdateRefreshToken updates the refresh token for a user
func (r *UserRepository) UpdateRefreshToken(ctx context.Context, userID, refreshToken string) error {
	expr, err := expression.NewBuilder().
		WithUpdate(expression.Set(
			expression.Name("refresh_token"),
			expression.Value(refreshToken)).
			Set(expression.Name("updated_at"),
				expression.Value(time.Now()))).
		Build()
	if err != nil {
		return err
	}

	_, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	return err
}

// emailExists checks if an email already exists in the database
func (r *UserRepository) emailExists(ctx context.Context, email string) (bool, error) {
	expr, err := expression.NewBuilder().
		WithFilter(expression.Name("email").Equal(expression.Value(email))).
		Build()
	if err != nil {
		return false, err
	}

	result, err := r.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String(r.tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Select:                    types.SelectCount,
	})
	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// GetCustomerByID retrieves a customer profile by ID
func (r *UserRepository) GetCustomerByID(ctx context.Context, id string) (*domain.Customer, error) {
    // Get the base user
    user, err := r.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Verify user is a customer
    if user.UserType != domain.UserTypeCustomer {
        return nil, errors.New("user is not a customer")
    }
    
    // Create customer with base user data
    customer := &domain.Customer{
        User: *user,
    }
    
    // Retrieve customer-specific attributes
    result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Extract customer-specific fields
    if result.Item != nil {
        if addr, ok := result.Item["address"]; ok {
            if addrValue, ok := addr.(*types.AttributeValueMemberS); ok {
                customer.Address = addrValue.Value
            }
        }
        
        if loyalty, ok := result.Item["loyalty_id"]; ok {
            if loyaltyValue, ok := loyalty.(*types.AttributeValueMemberS); ok {
                customer.LoyaltyID = loyaltyValue.Value
            }
        }
    }
    
    return customer, nil
}

// GetDriverByID retrieves a driver profile by ID
func (r *UserRepository) GetDriverByID(ctx context.Context, id string) (*domain.Driver, error) {
    // Get the base user
    user, err := r.GetUserByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Verify user is a driver
    if user.UserType != domain.UserTypeDriver {
        return nil, errors.New("user is not a driver")
    }
    
    // Create driver with base user data
    driver := &domain.Driver{
        User: *user,
    }
    
    // Retrieve driver-specific attributes
    result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
    })
    if err != nil {
        return nil, err
    }
    
    // Extract driver-specific fields
    if result.Item != nil {
        if license, ok := result.Item["license_number"]; ok {
            if licenseValue, ok := license.(*types.AttributeValueMemberS); ok {
                driver.LicenseNumber = licenseValue.Value
            }
        }
        
        if expiry, ok := result.Item["license_expiry"]; ok {
            if t, ok := expiry.(*types.AttributeValueMemberS); ok {
                expiryTime, err := time.Parse(time.RFC3339, t.Value)
                if err == nil {
                    driver.LicenseExpiry = expiryTime
                }
            }
        }
        
        if vehicle, ok := result.Item["vehicle_info"]; ok {
            if vehicleValue, ok := vehicle.(*types.AttributeValueMemberS); ok {
                driver.VehicleInfo = vehicleValue.Value
            }
        }
        
        if verified, ok := result.Item["is_verified"]; ok {
            if verifiedValue, ok := verified.(*types.AttributeValueMemberBOOL); ok {
                driver.IsVerified = verifiedValue.Value
            }
        }
    }
    
    return driver, nil
}

// UpdateCustomerProfile updates a customer's profile information
func (r *UserRepository) UpdateCustomerProfile(ctx context.Context, id string, update domain.CustomerProfileUpdate) error {
    // Check if user exists and is a customer
    user, err := r.GetUserByID(ctx, id)
    if err != nil {
        return err
    }
    
    if user.UserType != domain.UserTypeCustomer {
        return errors.New("user is not a customer")
    }
    
    // Build update expression
    updateExpr := expression.UpdateBuilder{}
    updateExpr = updateExpr.Set(expression.Name("updated_at"), expression.Value(time.Now()))
    
    // Add fields to update
    if update.FirstName != "" {
        updateExpr = updateExpr.Set(expression.Name("first_name"), expression.Value(update.FirstName))
    }
    
    if update.LastName != "" {
        updateExpr = updateExpr.Set(expression.Name("last_name"), expression.Value(update.LastName))
    }
    
    if update.Phone != "" {
        updateExpr = updateExpr.Set(expression.Name("phone"), expression.Value(update.Phone))
    }
    
    if update.Address != "" {
        updateExpr = updateExpr.Set(expression.Name("address"), expression.Value(update.Address))
    }
    
    // Build expression
    expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()
    if err != nil {
        return err
    }
    
    // Update item
    _, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression:          expr.Update(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
    })
    
    return err
}

func (r *UserRepository) UpdateDriverProfile(ctx context.Context, id string, update domain.DriverProfileUpdate) error {
    // Check if user exists and is a driver
    user, err := r.GetUserByID(ctx, id)
    if err != nil {
        return err
    }
    
    if user.UserType != domain.UserTypeDriver {
        return errors.New("user is not a driver")
    }
    
    // Build update expression
    updateExpr := expression.UpdateBuilder{}
    updateExpr = updateExpr.Set(expression.Name("updated_at"), expression.Value(time.Now()))
    
    // Add fields to update
    if update.FirstName != "" {
        updateExpr = updateExpr.Set(expression.Name("first_name"), expression.Value(update.FirstName))
    }
    
    if update.LastName != "" {
        updateExpr = updateExpr.Set(expression.Name("last_name"), expression.Value(update.LastName))
    }
    
    if update.Phone != "" {
        updateExpr = updateExpr.Set(expression.Name("phone"), expression.Value(update.Phone))
    }
    
    if update.VehicleInfo != "" {
        updateExpr = updateExpr.Set(expression.Name("vehicle_info"), expression.Value(update.VehicleInfo))
    }
    
    if update.LicenseNumber != "" {
        updateExpr = updateExpr.Set(expression.Name("license_number"), expression.Value(update.LicenseNumber))
    }
    
    if update.LicenseExpiry != "" {
        updateExpr = updateExpr.Set(expression.Name("license_expiry"), expression.Value(update.LicenseExpiry))
    }
    
    // Build expression
    expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()
    if err != nil {
        return err
    }
    
    // Update item
    _, err = r.client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
        TableName: aws.String(r.tableName),
        Key: map[string]types.AttributeValue{
            "id": &types.AttributeValueMemberS{Value: id},
        },
        UpdateExpression:          expr.Update(),
        ExpressionAttributeNames:  expr.Names(),
        ExpressionAttributeValues: expr.Values(),
    })
    
    return err
}
