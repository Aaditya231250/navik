package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// DriverProfile represents a driver's profile information
type DriverProfile struct {
	DriverID       string    `json:"driver_id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Status         string    `json:"status"`
	VehicleType    string    `json:"vehicle_type"`
	VehicleDetails struct {
		Make         string `json:"make"`
		Model        string `json:"model"`
		Year         int    `json:"year"`
		Color        string `json:"color"`
		LicensePlate string `json:"license_plate"`
	} `json:"vehicle_details"`
	LicenseNumber  string    `json:"license_number"`
	LicenseExpiry  time.Time `json:"license_expiry"`
	Documents      []string  `json:"documents"`
	Rating         float64   `json:"rating"`
	AccountDetails struct {
		BankName   string `json:"bank_name"`
		AccountNo  string `json:"account_no"`
		IFSCCode   string `json:"ifsc_code"`
		HolderName string `json:"holder_name"`
	} `json:"account_details"`
}

// driverProfiles is an in-memory store for driver profiles
// In a real application, this would be replaced with a database
var driverProfiles = make(map[string]DriverProfile)

// HandleDriverRoutes sets up the HTTP routes for driver profile management
func HandleDriverRoutes() {
	http.HandleFunc("/api/v1/drivers", handleDrivers)
	http.HandleFunc("/api/v1/drivers/", handleDriverByID)
}

// handleDrivers handles the creation of new driver profiles
func handleDrivers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		createDriverProfile(w, r)
		return
	}

	if r.Method == http.MethodGet {
		getAllDriverProfiles(w, r)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleDriverByID handles operations on a specific driver profile
func handleDriverByID(w http.ResponseWriter, r *http.Request) {
	// Extract driver ID from URL
	driverID := r.URL.Path[len("/api/v1/drivers/"):]
	
	if driverID == "" {
		http.Error(w, "Driver ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getDriverProfile(w, r, driverID)
	case http.MethodPut:
		updateDriverProfile(w, r, driverID)
	case http.MethodDelete:
		deleteDriverProfile(w, r, driverID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// createDriverProfile creates a new driver profile
func createDriverProfile(w http.ResponseWriter, r *http.Request) {
	var driver DriverProfile
	
	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&driver)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if driver.Name == "" || driver.Email == "" || driver.Phone == "" {
		http.Error(w, "Name, email, and phone are required fields", http.StatusBadRequest)
		return
	}
	
	// Check if driver already exists
	if _, exists := driverProfiles[driver.DriverID]; exists {
		http.Error(w, "Driver already exists", http.StatusConflict)
		return
	}
	
	// Set creation and update timestamps
	now := time.Now()
	driver.CreatedAt = now
	driver.UpdatedAt = now
	
	// Set default status if not provided
	if driver.Status == "" {
		driver.Status = "OFFLINE"
	}
	
	// Set default rating if not provided
	if driver.Rating == 0 {
		driver.Rating = 5.0
	}
	
	// Store the driver profile
	driverProfiles[driver.DriverID] = driver
	
	// Return the created profile
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(driver)
	
	log.Printf("Created driver profile: %s", driver.DriverID)
}

// getAllDriverProfiles returns all driver profiles
func getAllDriverProfiles(w http.ResponseWriter, r *http.Request) {
	// Convert map to slice for JSON response
	var drivers []DriverProfile
	for _, driver := range driverProfiles {
		drivers = append(drivers, driver)
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(drivers)
}

// getDriverProfile returns a specific driver profile
func getDriverProfile(w http.ResponseWriter, r *http.Request, driverID string) {
	driver, exists := driverProfiles[driverID]
	if !exists {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(driver)
}

// updateDriverProfile updates an existing driver profile
func updateDriverProfile(w http.ResponseWriter, r *http.Request, driverID string) {
	// Check if driver exists
	_, exists := driverProfiles[driverID]
	if !exists {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}
	
	var updatedDriver DriverProfile
	
	// Decode JSON request body
	err := json.NewDecoder(r.Body).Decode(&updatedDriver)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Ensure driver ID matches
	updatedDriver.DriverID = driverID
	
	// Preserve creation time
	updatedDriver.CreatedAt = driverProfiles[driverID].CreatedAt
	
	// Update timestamp
	updatedDriver.UpdatedAt = time.Now()
	
	// Update the driver profile
	driverProfiles[driverID] = updatedDriver
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedDriver)
	
	log.Printf("Updated driver profile: %s", driverID)
}

// deleteDriverProfile deletes a driver profile
func deleteDriverProfile(w http.ResponseWriter, r *http.Request, driverID string) {
	// Check if driver exists
	_, exists := driverProfiles[driverID]
	if !exists {
		http.Error(w, "Driver not found", http.StatusNotFound)
		return
	}
	
	// Delete the driver profile
	delete(driverProfiles, driverID)
	
	// Return success message
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message": "Driver profile deleted successfully"}`)
	
	log.Printf("Deleted driver profile: %s", driverID)
}
