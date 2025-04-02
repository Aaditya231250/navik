package model

import (
	"fmt"
	"time"
)

type UserLocation struct {
	UserID      string  `json:"user_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	RequestType string  `json:"request_type"`
}


func (l *UserLocation) Validate() error {
	if l.UserID == "" {
		return fmt.Errorf("driver_id is required")
	}
	if l.City == "" {
		return fmt.Errorf("city is required")
	}
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	if l.Timestamp == 0 {
		l.Timestamp = time.Now().Unix()
	}

	return nil
}

type EnrichedUserLocation struct {
	UserLocation
	H3Index9 string
	H3Index8 string
	H3Index7 string
}

// DriverLocation represents a driver's location from the database
type DriverLocation struct {
	// Base driver information
	DriverID    string    `json:"driver_id" dynamodbav:"DriverID"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Location    string    `json:"location" dynamodbav:"Location"`
	VehicleType string    `json:"vehicle_type" dynamodbav:"Vehicle"`
	Status      string    `json:"status" dynamodbav:"Status"`
	LastUpdated time.Time `json:"last_updated"`
	
	// H3 indices at different resolutions
	H3Res9      string    `json:"h3_res9" dynamodbav:"H3Res9"`
	H3Res8      string    `json:"h3_res8" dynamodbav:"H3Res8"`
	H3Res7      string    `json:"h3_res7" dynamodbav:"H3Res7"`
	
	// Matching-related fields (not stored in DB)
	Distance    float64   `json:"distance,omitempty"`
	ETA         int       `json:"eta_minutes,omitempty"`
}

// DriverResponse represents the formatted response to send back to the user
type DriverResponse struct {
	UserID      string       `json:"user_id"`
	RequestTime int64        `json:"request_time"`
	Drivers     []DriverInfo `json:"drivers"`
	Status      string       `json:"status"`
}

type DriverInfo struct {
	DriverID    string  `json:"driver_id"`
	VehicleType string  `json:"vehicle_type"`
	Distance    float64 `json:"distance_km"`
	ETA         int     `json:"eta_minutes"`
}
