// internal/handler/location.go
package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"location-service/internal/model"
	"location-service/internal/service"
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

	var loc model.Location
	if err := json.NewDecoder(r.Body).Decode(&loc); err != nil {
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

func (h *LocationHandler) HandleGetDriverLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	driverID := r.URL.Query().Get("driver_id")
	if driverID == "" {
		http.Error(w, "Missing driver_id parameter", http.StatusBadRequest)
		return
	}

	location, err := h.service.GetDriverLocation(r.Context(), driverID)
	if err != nil {
		log.Printf("Error getting driver location: %v", err)
		http.Error(w, "Failed to get driver location", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(location)
}

func (h *LocationHandler) HandleGetDriversByCity(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, "Missing city parameter", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	locations, err := h.service.GetDriversByCity(r.Context(), city, limit)
	if err != nil {
		log.Printf("Error getting drivers by city: %v", err)
		http.Error(w, "Failed to get drivers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func (h *LocationHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}

// SetupRoutes configures all HTTP routes
func (h *LocationHandler) SetupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/location", h.HandleLocationUpdate)
	mux.HandleFunc("/api/driver/location", h.HandleGetDriverLocation)
	mux.HandleFunc("/api/city/drivers", h.HandleGetDriversByCity)
	mux.HandleFunc("/health", h.HandleHealthCheck)
}
