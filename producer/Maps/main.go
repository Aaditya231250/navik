package main

import (
	"context"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
	"os"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Place represents a location suggestion
type Place struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
}

// PlaceDetails represents detailed information about a place
type PlaceDetails struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	PhoneNumber string  `json:"phoneNumber,omitempty"`
	Website     string  `json:"website,omitempty"`
	Rating      float64 `json:"rating,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	// Add more fields as needed
}

// SearchRequest represents the incoming search query
type SearchRequest struct {
	Type    string `json:"type"`
	Query   string `json:"query,omitempty"`
	PlaceID string `json:"placeId,omitempty"`
}

// SearchResponse represents the outgoing search results
type SearchResponse struct {
	Places []Place `json:"places,omitempty"`
}

// PlaceDetailsResponse represents the response with place details
type PlaceDetailsResponse struct {
	PlaceDetails PlaceDetails `json:"placeDetails"`
}

// Initialize Google Maps client and loggers
var mapsClient *maps.Client
var infoLog *log.Logger
var errorLog *log.Logger
var debugLog *log.Logger

func init() {
	// Set up loggers
	infoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLog = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		errorLog.Printf("Error loading .env file: %v", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("GMAPS_API_KEY")
	if apiKey == "" {
		errorLog.Println("GMAPS_API_KEY environment variable not set")
	}

	infoLog.Println("Initializing Google Maps client")
	mapsClient, err = maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		errorLog.Fatalf("Failed to create Google Maps client: %v", err)
		os.Exit(1)
	}
	infoLog.Println("Google Maps client initialized successfully")
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	infoLog.Printf("New connection request from %s", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		errorLog.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()
	infoLog.Printf("WebSocket connection established with %s", conn.RemoteAddr())

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			errorLog.Printf("Error reading message: %v", err)
			break
		}

		var searchReq SearchRequest
		if err := json.Unmarshal(message, &searchReq); err != nil {
			errorLog.Printf("Error unmarshalling request: %v", err)
			continue
		}

		// Check the type of request
		switch searchReq.Type {
		case "search":
			// Handle search query
			handleSearchQuery(conn, searchReq.Query)
		case "placeSelected":
			// Handle place selection
			handlePlaceSelection(conn, searchReq.PlaceID)
		default:
			errorLog.Printf("Unknown request type: %s", searchReq.Type)
		}
	}

	infoLog.Printf("Connection closed with %s", r.RemoteAddr)
}

func handleSearchQuery(conn *websocket.Conn, query string) {
	debugLog.Printf("Received search query: %s", query)
	var results []Place

	if query != "" {
		// Create autocomplete request
		r := &maps.PlaceAutocompleteRequest{
			Input: query,
			// Optional: Add location bias or restrictions if needed
			// Location: &maps.LatLng{Lat: 37.7749, Lng: -122.4194},
			// Radius: 50000, // meters
		}

		// Call the Places API
		startTime := time.Now()
		infoLog.Printf("Sending request to Google Maps API for query: %s", query)
		resp, err := mapsClient.PlaceAutocomplete(context.Background(), r)
		if err != nil {
			errorLog.Printf("Google Maps API error: %v", err)
		} else {
			debugLog.Printf("Google Maps API response time: %v", time.Since(startTime))
			debugLog.Printf("Received %d predictions from Google Maps API", len(resp.Predictions))

			// Convert Google Maps predictions to our Place format
			for _, prediction := range resp.Predictions {
				place := Place{
					ID:      prediction.PlaceID,
					Name:    prediction.StructuredFormatting.MainText,
					Address: prediction.Description,
				}
				results = append(results, place)
			}
		}
	}

	// Send the results back to the client
	response := SearchResponse{Places: results}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		errorLog.Printf("Error marshalling response: %v", err)
		return
	}

	infoLog.Printf("Sending %d results back to client", len(results))
	if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
		errorLog.Printf("Error writing message: %v", err)
		return
	}
	debugLog.Println("Response sent successfully")
}

func handlePlaceSelection(conn *websocket.Conn, placeID string) {
	infoLog.Printf("Selected place ID received: %s", placeID)

	// Create place details request
	r := &maps.PlaceDetailsRequest{
		PlaceID: placeID,
		// Include the fields you need
		Fields: []maps.PlaceDetailsFieldMask{
			maps.PlaceDetailsFieldMaskName,
			maps.PlaceDetailsFieldMaskGeometry,
		},
	}

	// Call the Places API for details
	startTime := time.Now()
	infoLog.Printf("Fetching place details for ID: %s", placeID)
	resp, err := mapsClient.PlaceDetails(context.Background(), r)
	if err != nil {
		errorLog.Printf("Google Maps Place Details API error: %v", err)
		return
	}

	debugLog.Printf("Google Maps Place Details API response time: %v", time.Since(startTime))

	// Convert API response to our PlaceDetails format
	details := PlaceDetails{
		ID:   resp.PlaceID,
		Name: resp.Name,
	}

	// Add location if available
	if resp.Geometry.Location.Lat != 0 || resp.Geometry.Location.Lng != 0 {
		details.Latitude = resp.Geometry.Location.Lat
		details.Longitude = resp.Geometry.Location.Lng
	}

	// Send the details back to the client
	response := PlaceDetailsResponse{PlaceDetails: details}
	responseJSON, err := json.Marshal(response)
	if err != nil {
		errorLog.Printf("Error marshalling place details response: %v", err)
		return
	}

	infoLog.Printf("Sending place details back to client for: %s", details.Name)
	if err := conn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
		errorLog.Printf("Error writing place details message: %v", err)
		return
	}
	debugLog.Println("Place details response sent successfully")
}

func main() {
	http.HandleFunc("/search", searchHandler)

	serverAddr := ":8080"
	infoLog.Printf("Starting WebSocket server on %s", serverAddr)
	infoLog.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		errorLog.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
