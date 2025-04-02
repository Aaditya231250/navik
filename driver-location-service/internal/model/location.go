package model

type Location struct {
	DriverID    string  `json:"driver_id"`
	City        string  `json:"city"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Timestamp   int64   `json:"timestamp"`
	VehicleType string  `json:"vehicle_type"`
	Status      string  `json:"status"`
}
