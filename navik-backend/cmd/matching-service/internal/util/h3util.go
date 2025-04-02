package util

import (
	"github.com/uber/h3-go/v3"
)

// GeoToH3Index converts lat/lng to H3 index string
func GeoToH3Index(lat, lng float64, resolution int) string {
	h3Index := h3.FromGeo(h3.GeoCoord{
		Latitude:  lat,
		Longitude: lng,
	}, resolution)
	
	return h3.ToString(h3Index)
}

// GetH3Neighbors returns the neighboring H3 cells for a given H3 index
func GetH3Neighbors(h3Index string) []string {
	// Convert string index to H3 index
	h3IndexInt := h3.FromString(h3Index)
	
	// Get the k-ring of neighbors (k=1 means immediate neighbors)
	neighbors := h3.KRing(h3IndexInt, 1)
	
	// Convert to strings and filter out the center cell (which is the original cell)
	var neighborStrings []string
	for _, neighbor := range neighbors {
		neighborString := h3.ToString(neighbor)
		if neighborString != h3Index {
			neighborStrings = append(neighborStrings, neighborString)
		}
	}
	
	return neighborStrings
}

// SafePrefix returns a safe prefix of the string
func SafePrefix(s string, length int) string {
	if len(s) < length {
		return s
	}
	return s[:length]
}
