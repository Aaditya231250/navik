package models

// Route represents a dummy structure for a route
type Route struct {
	ID          string  `json:"id"`
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	Distance    float64 `json:"distance"`
}

// NewRoute creates a new dummy route
func NewRoute(id, source, destination string, distance float64) *Route {
	return &Route{
		ID:          id,
		Source:      source,
		Destination: destination,
		Distance:    distance,
	}
}
