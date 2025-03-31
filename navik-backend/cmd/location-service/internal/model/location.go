// internal/model/location.go
package model

import (
	"fmt"
	"time"
)

type Location struct {
	DriverID  string  `json:"driver_id"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
	Status    string  `json:"status,omitempty"` // Available, Busy, Offline
	Speed     float64 `json:"speed,omitempty"`
	Heading   float64 `json:"heading,omitempty"` // Direction in degrees
}

// Validate checks if the location data is valid
func (l *Location) Validate() error {
	if l.DriverID == "" {
		return fmt.Errorf("driver_id is required")
	}
	if l.City == "" {
		return fmt.Errorf("city is required")
	}
	// Basic latitude/longitude validation
	if l.Latitude < -90 || l.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if l.Longitude < -180 || l.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}

	// Set default timestamp if not provided
	if l.Timestamp == 0 {
		l.Timestamp = time.Now().Unix()
	}

	return nil
}
