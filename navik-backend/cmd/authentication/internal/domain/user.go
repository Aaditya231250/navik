package domain

import (
    "time"
)

// UserType defines the type of user in the system
type UserType string

const (
    UserTypeCustomer UserType = "customer"
    UserTypeDriver   UserType = "driver"
)

// User represents common user attributes
type User struct {
    ID           string    `json:"id" dynamodbav:"id"`
    Email        string    `json:"email" dynamodbav:"email"`
    Password     string    `json:"-" dynamodbav:"password"`
    UserType     UserType  `json:"user_type" dynamodbav:"user_type"`
    Phone        string    `json:"phone" dynamodbav:"phone"`
    FirstName    string    `json:"first_name" dynamodbav:"first_name"`
    LastName     string    `json:"last_name" dynamodbav:"last_name"`
    CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
    UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`
    IsActive     bool      `json:"is_active" dynamodbav:"is_active"`
    RefreshToken string    `json:"-" dynamodbav:"refresh_token,omitempty"`
}

// Customer contains customer-specific fields
type Customer struct {
    User
    Address    string `json:"address,omitempty" dynamodbav:"address,omitempty"`
    LoyaltyID  string `json:"loyalty_id,omitempty" dynamodbav:"loyalty_id,omitempty"`
}

// Driver contains driver-specific fields
type Driver struct {
    User
    LicenseNumber string    `json:"license_number,omitempty" dynamodbav:"license_number,omitempty"`
    LicenseExpiry time.Time `json:"license_expiry,omitempty" dynamodbav:"license_expiry,omitempty"`
    VehicleInfo   string    `json:"vehicle_info,omitempty" dynamodbav:"vehicle_info,omitempty"`
    IsVerified    bool      `json:"is_verified" dynamodbav:"is_verified"`
}

// Registration represents data needed for user registration
type Registration struct {
    Email       string   `json:"email" validate:"required,email"`
    Password    string   `json:"password" validate:"required,min=8"`
    UserType    UserType `json:"user_type" validate:"required,oneof=customer driver"`
    Phone       string   `json:"phone" validate:"required"`
    FirstName   string   `json:"first_name" validate:"required"`
    LastName    string   `json:"last_name" validate:"required"`
    
    // Optional fields based on user type
    Address         string    `json:"address,omitempty"`
    LicenseNumber   string    `json:"license_number,omitempty"`
    LicenseExpiry   time.Time `json:"license_expiry,omitempty"`
    VehicleInfo     string    `json:"vehicle_info,omitempty"`
}

// Login represents data needed for user login
type Login struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required"`
}

// AuthResponse represents authentication response with tokens
type AuthResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"` 
    UserID       string `json:"user_id"`
    UserType     string `json:"user_type"`
}

type PasswordReset struct {
    Email       string `json:"email" validate:"required,email"`
    ResetToken  string `json:"reset_token,omitempty"`
    NewPassword string `json:"new_password,omitempty" validate:"omitempty,min=8"`
}
