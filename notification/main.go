package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// DriverNotification represents a notification to be sent to a driver
type DriverNotification struct {
	Type        string  `json:"type"`
	UserID      string  `json:"user_id"`
	DriverID    string  `json:"driver_id"`
	Priority    int     `json:"priority"`
	PickupLat   float64 `json:"pickup_lat"`
	PickupLong  float64 `json:"pickup_long"`
	Distance    float64 `json:"distance_km"`
	ETA         int     `json:"eta_minutes"`
	RequestTime int64   `json:"request_time"`
	ExpiresAt   int64   `json:"expires_at"`
}

// DriverResponse represents a driver's response to a ride request
type DriverResponse struct {
	Type        string `json:"type"`
	UserID      string `json:"user_id"`
	DriverID    string `json:"driver_id"`
	RequestTime int64  `json:"request_time"`
	ResponseTime int64 `json:"response_time"`
	Status      string `json:"status"` // "ACCEPT" or "REJECT"
}

// DriverConnection represents a connected driver
type DriverConnection struct {
	DriverID     string
	Conn         *websocket.Conn
	LastActivity time.Time
	Status       string // "AVAILABLE", "BUSY", "OFFLINE"
}

var (
	// Map to store active driver connections
	driverConnections = make(map[string]*DriverConnection)
	connectionsMutex  sync.RWMutex
	
	// WebSocket upgrader
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins in development
		},
	}

	// Channel for pending notifications that need user input
	pendingNotifications = make(chan DriverNotification, 100)
	
	// Connection to matching service
	matchingServiceConn *websocket.Conn
)

func main() {
	// Start a goroutine to handle terminal input for driver responses
	go handleTerminalInput()
	
	// Driver connection endpoint
	http.HandleFunc("/ws/driver", handleDriverConnection)
	
	// Matching service connection endpoint
	http.HandleFunc("/ws/matching", handleMatchingServiceConnection)
	
	log.Println("Notification service started on :9080")
	if err := http.ListenAndServe(":9080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// handleTerminalInput processes user input from the terminal to simulate driver responses
func handleTerminalInput() {
	scanner := bufio.NewScanner(os.Stdin)
	
	fmt.Println("=== Driver Notification Terminal ===")
	fmt.Println("Waiting for ride requests...")
	
	// Process pending notifications
	for notification := range pendingNotifications {
		// Display the notification
		fmt.Printf("\n\n=== New Ride Request (Priority: %d) ===\n", notification.Priority)
		fmt.Printf("Driver ID: %s\n", notification.DriverID)
		fmt.Printf("User ID: %s\n", notification.UserID)
		fmt.Printf("Pickup Location: %.6f, %.6f\n", notification.PickupLat, notification.PickupLong)
		fmt.Printf("Distance: %.2f km\n", notification.Distance)
		fmt.Printf("ETA: %d minutes\n", notification.ETA)
		fmt.Printf("Request Time: %s\n", time.Unix(notification.RequestTime, 0).Format(time.RFC3339))
		fmt.Printf("Expires At: %s\n", time.Unix(notification.ExpiresAt, 0).Format(time.RFC3339))
		
		// Prompt for response
		fmt.Printf("\nDo you want driver %s to accept this ride? (y/n): ", notification.DriverID)
		
		if scanner.Scan() {
			response := strings.ToLower(strings.TrimSpace(scanner.Text()))
			
			// Create driver response
			driverResponse := DriverResponse{
				Type:         "RIDE_RESPONSE",
				UserID:       notification.UserID,
				DriverID:     notification.DriverID,
				RequestTime:  notification.RequestTime,
				ResponseTime: time.Now().Unix(),
			}
			
			if response == "y" || response == "yes" {
				driverResponse.Status = "ACCEPT"
				fmt.Printf("Driver %s ACCEPTED the ride request\n", notification.DriverID)
				
				// Update driver status
				connectionsMutex.Lock()
				if conn, exists := driverConnections[notification.DriverID]; exists {
					conn.Status = "BUSY"
				}
				connectionsMutex.Unlock()
			} else {
				driverResponse.Status = "REJECT"
				fmt.Printf("Driver %s REJECTED the ride request\n", notification.DriverID)
			}
			
			// Send response to matching service if connected
			if matchingServiceConn != nil {
				responseJSON, _ := json.Marshal(driverResponse)
				if err := matchingServiceConn.WriteMessage(websocket.TextMessage, responseJSON); err != nil {
					log.Printf("Error sending response to matching service: %v", err)
				}
			}
		}
		
		fmt.Println("\nWaiting for next ride request...")
	}
}

// handleDriverConnection manages WebSocket connections from drivers
func handleDriverConnection(w http.ResponseWriter, r *http.Request) {
	// Extract driver ID from request (using auth token/query param)
	driverID := r.URL.Query().Get("driver_id")
	if driverID == "" {
		http.Error(w, "Missing driver_id parameter", http.StatusBadRequest)
		return
	}
	
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	
	log.Printf("Driver %s connected", driverID)
	fmt.Printf("\nDriver %s connected to the system\n", driverID)
	
	// Register the driver connection
	connectionsMutex.Lock()
	driverConnections[driverID] = &DriverConnection{
		DriverID:     driverID,
		Conn:         conn,
		LastActivity: time.Now(),
		Status:       "AVAILABLE",
	}
	connectionsMutex.Unlock()
	
	// Handle WebSocket lifecycle
	defer func() {
		conn.Close()
		connectionsMutex.Lock()
		delete(driverConnections, driverID)
		connectionsMutex.Unlock()
		log.Printf("Driver %s disconnected", driverID)
		fmt.Printf("\nDriver %s disconnected from the system\n", driverID)
	}()
	
	// Keep connection alive and handle driver status updates
	for {
		// Read messages from driver (status updates, etc.)
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
		
		// Process driver messages (status updates, ride responses)
		var update map[string]interface{}
		if err := json.Unmarshal(message, &update); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}
		
		// Update driver status if provided
		if status, ok := update["status"].(string); ok {
			connectionsMutex.Lock()
			if conn, exists := driverConnections[driverID]; exists {
				conn.Status = status
				conn.LastActivity = time.Now()
			}
			connectionsMutex.Unlock()
			log.Printf("Driver %s status updated to %s", driverID, status)
			fmt.Printf("\nDriver %s status updated to %s\n", driverID, status)
		}
	}
}

// handleMatchingServiceConnection manages WebSocket connection from the matching service
func handleMatchingServiceConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	
	log.Printf("Matching service connected")
	fmt.Println("\nMatching service connected")
	
	// Store the connection for later use
	matchingServiceConn = conn
	
	// Handle WebSocket lifecycle
	defer func() {
		conn.Close()
		matchingServiceConn = nil
		log.Printf("Matching service disconnected")
		fmt.Println("\nMatching service disconnected")
	}()
	
	// Process messages from matching service
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			break
		}
		
		// Process driver notifications from matching service
		var notifications []DriverNotification
		if err := json.Unmarshal(message, &notifications); err != nil {
			log.Printf("Error unmarshaling notifications: %v", err)
			continue
		}
		
		fmt.Printf("\nReceived %d driver notifications from matching service\n", len(notifications))
		
		// Send notifications to terminal for user input
		for _, notification := range notifications {
			// Only process notifications for drivers that are connected and available
			connectionsMutex.RLock()
			driverConn, exists := driverConnections[notification.DriverID]
			isAvailable := exists && driverConn.Status == "AVAILABLE"
			connectionsMutex.RUnlock()
			
			if !isAvailable {
				log.Printf("Driver %s not available for notification", notification.DriverID)
				continue
			}
			
			// Send to terminal handler
			pendingNotifications <- notification
		}
	}
}
