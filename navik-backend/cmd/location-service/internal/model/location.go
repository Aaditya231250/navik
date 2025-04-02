package model

import (
	"fmt"
	"time"
)

type Location struct {
	DriverID    string  `json:"driver_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	VehicleType string  `json:"vehicle_type"`
	Status      string  `json:"status"`
}

type LocationDB struct {
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
	VehicleType   string `json:"vehicle_type" dynamodbav:"vehicle_type"`
	Status        string `json:"status" dynamodbav:"status"`
	UpdatedAt     int64  `json:"updated_at" dynamodbav:"updated_at"`
	ExpiresAt     int64  `json:"expires_at" dynamodbav:"expires_at"`
}

func (l *Location) Validate() error {
	if l.DriverID == "" {
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

	if l.VehicleType == "" {
		return fmt.Errorf("vehicle_type is required")
	}

	if l.Status == "" {
		return fmt.Errorf("status is required")
	}

	if l.Timestamp == 0 {
		l.Timestamp = time.Now().Unix()
	}

	return nil
}
