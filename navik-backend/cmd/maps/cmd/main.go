package main

import (
	"log"
	"net/http"
	"os"

	"internal-maps/internal/handlers"
	"internal-maps/internal/utils"
)

func main() {
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLog := log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	_, err := utils.LoadConfig()
	if err != nil {
		errorLog.Fatalf("Failed to load configuration: %v", err)
	}

	wsHandler := handlers.NewWebSocketHandler(infoLog, errorLog, debugLog)

	http.HandleFunc("/search", wsHandler.HandleSearch)

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	serverAddr := ":8080"
	infoLog.Printf("Starting WebSocket server on %s", serverAddr)
	infoLog.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		errorLog.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
