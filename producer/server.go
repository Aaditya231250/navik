package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type DriverLocationUpdate struct {
	DriverID  string  `json:"driver_id,omitempty"`
	RiderID   string  `json:"rider_id,omitempty"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timestamp int64   `json:"timestamp"`
}

type KafkaMessage struct {
	Topic   string      `json:"topic"`
	Payload interface{} `json:"payload"`
}

func publishHandler(w http.ResponseWriter, r *http.Request) {
	var locationUpdate DriverLocationUpdate
	if err := json.NewDecoder(r.Body).Decode(&locationUpdate); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var topic string
	if locationUpdate.DriverID != "" {
		topic = DriverTopic
	} else if locationUpdate.RiderID != "" {
		topic = RiderTopic
	} else {
		http.Error(w, "Either driver_id or rider_id is required", http.StatusBadRequest)
		return
	}

	if locationUpdate.Timestamp == 0 {
		locationUpdate.Timestamp = time.Now().Unix()
	}

	message := KafkaMessage{
		Topic:   topic,
		Payload: locationUpdate,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Failed to encode message", http.StatusInternalServerError)
		return
	}

	if err := publishMessage([]string{KafkaBroker}, topic, jsonData); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Location update published successfully"))
}

func StartServer() {
	http.HandleFunc("/api/driver/update-position", publishHandler)

	log.Printf("🚀 Server started on %s\n", ServerAddress)
	log.Fatal(http.ListenAndServe(ServerAddress, nil))
}