package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"navik/producer/maps/internal/models"
	"navik/producer/maps/internal/services"
	"navik/producer/maps/internal/utils"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	upgrader         websocket.Upgrader
	geocodingService *services.GeocodingService
	detailsService   *services.PlaceDetailsService
	infoLog          *log.Logger
	errorLog         *log.Logger
	debugLog         *log.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(infoLog, errorLog, debugLog *log.Logger) *WebSocketHandler {
	// Load configuration
	config, err := utils.LoadConfig()
	if err != nil {
		errorLog.Fatalf("Failed to load configuration: %v", err)
	}

	// Create services
	geocodingService := services.NewGeocodingService(config.MapsClient, infoLog, errorLog, debugLog)
	detailsService := services.NewPlaceDetailsService(config.MapsClient, infoLog, errorLog, debugLog)

	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		geocodingService: geocodingService,
		detailsService:   detailsService,
		infoLog:          infoLog,
		errorLog:         errorLog,
		debugLog:         debugLog,
	}
}

// HandleSearch handles WebSocket connections for place search
func (h *WebSocketHandler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	h.infoLog.Printf("New connection request from %s", r.RemoteAddr)
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.errorLog.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()
	h.infoLog.Printf("WebSocket connection established with %s", conn.RemoteAddr())

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			h.errorLog.Printf("Error reading message: %v", err)
			break
		}

		var searchReq models.SearchRequest
		if err := json.Unmarshal(message, &searchReq); err != nil {
			h.errorLog.Printf("Error unmarshalling request: %v", err)
			continue
		}

		// Check the type of request
		switch searchReq.Type {
		case "search":
			h.handleSearchQuery(conn, searchReq.Query)
		case "placeSelected":
			h.handlePlaceSelection(conn, searchReq.PlaceID)
		default:
			h.errorLog.Printf("Unknown request type: %s", searchReq.Type)
		}
	}

	h.infoLog.Printf("Connection closed with %s", r.RemoteAddr)
}

// handleSearchQuery processes search queries
func (h *WebSocketHandler) handleSearchQuery(conn *websocket.Conn, query string) {
	results, err := h.geocodingService.SearchPlaces(query)
	if err != nil {
		h.errorLog.Printf("Error searching places: %v", err)
		return
	}

	// Send the results back to the client
	response := models.SearchResponse{Places: results}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		h.errorLog.Printf("Error marshalling response: %v", err)
		return
	}

	h.infoLog.Printf("Sending %d results back to client", len(results))
	if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
		h.errorLog.Printf("Error writing message: %v", err)
		return
	}
	h.debugLog.Println("Response sent successfully")
}

// handlePlaceSelection processes place selection
func (h *WebSocketHandler) handlePlaceSelection(conn *websocket.Conn, placeID string) {
	h.infoLog.Printf("Selected place ID received: %s", placeID)

	details, err := h.detailsService.GetPlaceDetails(placeID)
	if err != nil {
		h.errorLog.Printf("Error getting place details: %v", err)
		return
	}

	// Send the details back to the client
	response := models.PlaceDetailsResponse{PlaceDetails: *details}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		h.errorLog.Printf("Error marshalling place details response: %v", err)
		return
	}

	h.infoLog.Printf("Sending place details back to client for: %s", details.Name)
	if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
		h.errorLog.Printf("Error writing place details message: %v", err)
		return
	}
	h.debugLog.Println("Place details response sent successfully")
}
