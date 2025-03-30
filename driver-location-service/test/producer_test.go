package test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"driver-location-service/internal/model"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
)

func MockSendLocationHandler(producer sarama.SyncProducer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var loc model.Location
		if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate required fields
		if loc.DriverID == "" || loc.City == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		// Send to Kafka
		jsonData, _ := json.Marshal(loc)
		topic := loc.City + "-locations"
		msg := &sarama.ProducerMessage{
			Topic: topic,
			Value: sarama.ByteEncoder(jsonData),
		}

		if _, _, err := producer.SendMessage(msg); err != nil {
			http.Error(w, "Failed to process location update", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "success",
			"message": "Location update processed",
		})
	}
}

func TestSendLocation(t *testing.T) {
	// Create a mock producer
	mockProducer := mocks.NewSyncProducer(t, nil)
	mockProducer.ExpectSendMessageAndSucceed()

	// Create a test handler with the mock producer
	handler := MockSendLocationHandler(mockProducer)

	// Create test location data
	testLocation := model.Location{
		DriverID:    "MH-12345678",
		City:        "mumbai",
		Latitude:    19.076,
		Longitude:   72.877,
		Timestamp:   1647860964,
		VehicleType: "STANDARD",
		Status:      "ACTIVE",
	}

	// Create request body
	reqBody, _ := json.Marshal(testLocation)
	req, err := http.NewRequest("POST", "/api/location", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder and run handler
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
