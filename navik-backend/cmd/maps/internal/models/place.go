package models

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
