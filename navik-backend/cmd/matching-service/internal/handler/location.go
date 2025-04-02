package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"matching-service/internal/model"
	"matching-service/internal/service"
)

type LocationHandler struct {
	service service.LocationService
}

func NewLocationHandler(service service.LocationService) *LocationHandler {
	return &LocationHandler{
		service: service,
	}
}

func (h *LocationHandler) HandleLocationUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var loc model.UserLocation
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateLocation(r.Context(), loc); err != nil {
		log.Printf("Error updating location: %v", err)
		http.Error(w, "Failed to process location update", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Location update processed",
	})
}

func (h *LocationHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

func (h *LocationHandler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/matching", h.HandleLocationUpdate)
	mux.HandleFunc("/health", h.HandleHealthCheck)
}
