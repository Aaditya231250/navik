package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// MessageRequest represents the expected JSON request body
type MessageRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// publishHandler handles the POST request to send messages to Kafka
func publishHandler(w http.ResponseWriter, r *http.Request) {
	var msgReq MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&msgReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	brokers := []string{"localhost:9092"}
	topic := "driver-position-update"

	// Publish message to Kafka
	if err := publishMessage(brokers, topic, msgReq.Key, msgReq.Value); err != nil {
		http.Error(w, "Failed to publish message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Message published successfully"))
}

func StartServer() {
	http.HandleFunc("/api/driver/update-position", publishHandler)

	log.Println("Server started on :6969")
	log.Fatal(http.ListenAndServe(":6969", nil))
}
