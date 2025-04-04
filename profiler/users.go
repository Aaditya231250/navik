package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// UserProfile represents a user's profile information
type UserProfile struct {
	UserID         string    `json:"user_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Status         string    `json:"status"`
	PaymentMethods []struct {
		Type    string `json:"type"`
		Details string `json:"details"`
		Default bool   `json:"default"`
	} `json:"payment_methods"`
	HomeAddress struct {
		Street     string  `json:"street"`
		City       string  `json:"city"`
		State      string  `json:"state"`
		PostalCode string  `json:"postal_code"`
		Country    string  `json:"country"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
	} `json:"home_address"`
	WorkAddress struct {
		Street     string  `json:"street"`
		City       string  `json:"city"`
		State      string  `json:"state"`
		PostalCode string  `json:"postal_code"`
		Country    string  `json:"country"`
		Latitude   float64 `json:"latitude"`
		Longitude  float64 `json:"longitude"`
	} `json:"work_address"`
	Preferences map[string]interface{} `json:"preferences"`
}

// userProfiles is an in-memory store for user profiles
// In a real application, this would be replaced with a database
var userProfiles = make(map[string]UserProfile)

// HandleUserRoutes sets up the HTTP routes for user profile management
func HandleUserRoutes() {
	http.HandleFunc("/api/v1/users", handleUsers)
	http.HandleFunc("/api/v1/users/", handleUserByID)
}

// handleUsers handles the creation of new user profiles
func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createUserProfile(w, r)
		return
	}

	if r.Method == http.MethodGet {
		getAllUserProfiles(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleUserByID handles operations on a specific user profile
func handleUserByID(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL
	userID := r.URL.Path[len("/api/v1/users/"):]
	
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getUserProfile(w, r, userID)
	case http.MethodPut:
		updateUserProfile(w, r, userID)
	case http.MethodDelete:
		deleteUserProfile(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createUserProfile creates a new user profile
func createUserProfile(w http.ResponseWriter, r *http.Request) {
	var user UserProfile
	
	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if user.Name == "" || user.Email == "" || user.Phone == "" {
		http.Error(w, "Name, email, and phone are required fields", http.StatusBadRequest)
		return
	}
	
	// Check if user already exists
	if _, exists := userProfiles[user.UserID]; exists {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}
	
	// Set creation and update timestamps
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	
	// Set default status if not provided
	if user.Status == "" {
		user.Status = "ACTIVE"
	}
	
	// Initialize preferences if not provided
	if user.Preferences == nil {
		user.Preferences = make(map[string]interface{})
	}
	
	// Store the user profile
	userProfiles[user.UserID] = user
	
	// Return the created profile
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
	
	log.Printf("Created user profile: %s", user.UserID)
}

// getAllUserProfiles returns all user profiles
func getAllUserProfiles(w http.ResponseWriter, r *http.Request) {
	// Convert map to slice for JSON response
	var users []UserProfile
	for _, user := range userProfiles {
		users = append(users, user)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// getUserProfile returns a specific user profile
func getUserProfile(w http.ResponseWriter, r *http.Request, userID string) {
	user, exists := userProfiles[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// updateUserProfile updates an existing user profile
func updateUserProfile(w http.ResponseWriter, r *http.Request, userID string) {
	// Check if user exists
	_, exists := userProfiles[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	var updatedUser UserProfile
	
	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&updatedUser)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Ensure user ID matches
	updatedUser.UserID = userID
	
	// Preserve creation time
	updatedUser.CreatedAt = userProfiles[userID].CreatedAt
	
	// Update timestamp
	updatedUser.UpdatedAt = time.Now()
	
	// Update the user profile
	userProfiles[userID] = updatedUser
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
	
	log.Printf("Updated user profile: %s", userID)
}

// deleteUserProfile deletes a user profile
func deleteUserProfile(w http.ResponseWriter, r *http.Request, userID string) {
	// Check if user exists
	_, exists := userProfiles[userID]
	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	
	// Delete the user profile
	delete(userProfiles, userID)
	
	// Return success message
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "User profile deleted successfully"}`)
	
	log.Printf("Deleted user profile: %s", userID)
}

// Main function to start the server
func main() {
	// Set up routes
	HandleUserRoutes()
	HandleDriverRoutes()
	
	// Start the server
	log.Println("Profile service started on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
