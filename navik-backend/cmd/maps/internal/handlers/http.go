package handlers

import (
	"net/http"
)

// DummyHandler is a placeholder HTTP handler
func DummyHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, this is a dummy handler!"))
}
