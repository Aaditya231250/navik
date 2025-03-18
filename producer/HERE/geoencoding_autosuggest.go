package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	apiKey          = "Il21WEIs4ExhGhcXb7vxNILf_6boi44VGKORoJ3zqLE" // Replace with your actual HERE API key
	geocodeEndpoint = "https://geocode.search.hereapi.com/v1/geocode"
	suggestEndpoint = "https://autosuggest.search.hereapi.com/v1/autosuggest"
	requestTimeout  = 10 * time.Second
)

// Position represents geographic coordinates
type Position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Address represents a structured address
type Address struct {
	Label       string `json:"label"`
	CountryCode string `json:"countryCode"`
	CountryName string `json:"countryName,omitempty"`
	StateCode   string `json:"stateCode,omitempty"`
	State       string `json:"state,omitempty"`
	County      string `json:"county,omitempty"`
	City        string `json:"city,omitempty"`
	District    string `json:"district,omitempty"`
	Street      string `json:"street,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	HouseNumber string `json:"houseNumber,omitempty"`
}

// GeocodeResult represents a single geocoding result
type GeocodeResult struct {
	Title      string   `json:"title"`
	ID         string   `json:"id"`
	ResultType string   `json:"resultType"`
	Address    Address  `json:"address"`
	Position   Position `json:"position"`
	MapView    struct {
		West  float64 `json:"west"`
		South float64 `json:"south"`
		East  float64 `json:"east"`
		North float64 `json:"north"`
	} `json:"mapView"`
}

// GeocodeResponse represents the response from the geocode API
type GeocodeResponse struct {
	Items []GeocodeResult `json:"items"`
}

// SuggestResult represents a single autosuggest result
type SuggestResult struct {
	Title      string   `json:"title"`
	ID         string   `json:"id"`
	ResultType string   `json:"resultType"`
	Address    Address  `json:"address,omitempty"`
	Position   Position `json:"position,omitempty"`
	Highlights struct {
		Title   []struct{ Start, End int } `json:"title,omitempty"`
		Address []struct{ Start, End int } `json:"address,omitempty"`
	} `json:"highlights,omitempty"`
}

// SuggestResponse represents the response from the autosuggest API
type SuggestResponse struct {
	Items []SuggestResult `json:"items"`
}

// HEREClient is a client for the HERE Geocoding & Search API
type HEREClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewHEREClient creates a new HERE API client
func NewHEREClient(apiKey string) *HEREClient {
	return &HEREClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// makeRequest makes an HTTP request with GZIP compression
func (c *HEREClient) makeRequest(endpoint string, params url.Values) ([]byte, error) {
	// Add API key to parameters
	params.Add("apiKey", c.apiKey)

	// Create request
	req, err := http.NewRequest("GET", endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add headers for GZIP compression
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Add("User-Agent", "HEREGoClient (gzip)")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %s", resp.Status)
	}

	// Handle response body based on content encoding
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error creating gzip reader: %w", err)
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	// Read response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

// Geocode converts an address to coordinates
func (c *HEREClient) Geocode(address string) (*GeocodeResponse, error) {
	// Prepare parameters
	params := url.Values{}
	params.Add("q", address)

	// Make request
	body, err := c.makeRequest(geocodeEndpoint, params)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response GeocodeResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing geocode response: %w", err)
	}

	return &response, nil
}

// Autosuggest provides suggestions for incomplete addresses
func (c *HEREClient) Autosuggest(partialAddress string) (*SuggestResponse, error) {
	// Skip empty input
	if strings.TrimSpace(partialAddress) == "" {
		return &SuggestResponse{Items: []SuggestResult{}}, nil
	}

	// Prepare parameters
	params := url.Values{}
	params.Add("q", partialAddress)
	params.Add("resultTypes", "address,place")
	params.Add("limit", "5") // Limit to 5 suggestions for better UX

	// Make request
	body, err := c.makeRequest(suggestEndpoint, params)
	if err != nil {
		return nil, err
	}

	// Parse response
	var response SuggestResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing autosuggest response: %w", err)
	}

	return &response, nil
}

// AddressSearchFlow implements the real-world flow of autosuggest + geocoding
func (c *HEREClient) AddressSearchFlow() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Start typing an address (press Enter after each keystroke to simulate real-time suggestions):")

	var input string
	var selectedAddress string

	// Step 1: Autosuggest as user types
	for {
		fmt.Print("> ")
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println("Error reading input:", err)
			return
		}

		// Handle backspace
		if char == '\b' || char == 127 { // Backspace or Delete
			if len(input) > 0 {
				input = input[:len(input)-1]
			}
		} else if char == '\n' || char == '\r' { // Enter key
			// Check if user wants to finish typing
			if len(input) > 0 {
				fmt.Print("Continue typing or select a suggestion (1-5) or type 'done' to finish: ")
				selection, _ := reader.ReadString('\n')
				selection = strings.TrimSpace(selection)

				if selection == "done" {
					selectedAddress = input
					break
				}

				// Check if user selected a suggestion
				if n, err := fmt.Sscanf(selection, "%d", new(int)); err == nil && n > 0 {
					index, _ := fmt.Sscanf(selection, "%d", new(int))

					// Get suggestions again to have them available
					suggestions, err := c.Autosuggest(input)
					if err != nil {
						fmt.Println("Error getting suggestions:", err)
						continue
					}

					if index <= len(suggestions.Items) {
						selectedAddress = suggestions.Items[index-1].Title
						fmt.Printf("Selected: %s\n", selectedAddress)
						break
					} else {
						fmt.Println("Invalid selection, please continue typing or select a valid suggestion")
					}
				}
			}
		} else {
			// Append character to input
			input += string(char)
		}

		fmt.Printf("Current input: %s\n", input)

		// Get suggestions for current input
		if len(input) >= 2 { // Only suggest after at least 2 characters
			suggestions, err := c.Autosuggest(input)
			if err != nil {
				fmt.Println("Error getting suggestions:", err)
				continue
			}

			if len(suggestions.Items) > 0 {
				fmt.Println("Suggestions:")
				for i, item := range suggestions.Items {
					fmt.Printf("%d. %s\n", i+1, item.Title)
				}
			} else {
				fmt.Println("No suggestions found")
			}
		}
	}

	// Step 2: Geocode the selected address
	fmt.Printf("\nGeocoding selected address: %s\n", selectedAddress)
	response, err := c.Geocode(selectedAddress)
	if err != nil {
		fmt.Printf("Error geocoding address: %v\n", err)
		return
	}

	if len(response.Items) > 0 {
		result := response.Items[0]
		fmt.Printf("\nGeocoding Results:\n")
		fmt.Printf("Address: %s\n", result.Title)
		fmt.Printf("Coordinates: %.6f, %.6f\n", result.Position.Lat, result.Position.Lng)
		fmt.Printf("Type: %s\n", result.ResultType)

		if result.Address.City != "" {
			fmt.Printf("City: %s\n", result.Address.City)
		}
		if result.Address.State != "" {
			fmt.Printf("State: %s\n", result.Address.State)
		}
		if result.Address.CountryName != "" {
			fmt.Printf("Country: %s\n", result.Address.CountryName)
		}
	} else {
		fmt.Println("No geocoding results found")
	}
}

// Interactive demo that shows both the autosuggest and geocoding steps
func (c *HEREClient) RunInteractiveDemo() {
	fmt.Println("HERE Geocoding & Search API Demo")
	fmt.Println("--------------------------------")
	fmt.Println("This demo shows the real-world flow of address input:")
	fmt.Println("1. As you type, you'll get autosuggestions")
	fmt.Println("2. Once you select a complete address, it will be geocoded to coordinates")
	fmt.Println("--------------------------------")

	c.AddressSearchFlow()
}

func main() {
	// Create HERE API client
	client := NewHEREClient(apiKey)

	// Run the interactive demo
	client.RunInteractiveDemo()
}
