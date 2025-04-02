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
