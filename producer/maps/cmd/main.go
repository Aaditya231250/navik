package main

import (
	"log"
	"net/http"
	"os"

	"navik/producer/maps/internal/handlers"
	"navik/producer/maps/internal/utils"
)

func main() {
	// Set up loggers
	infoLog := log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLog := log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLog := log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Initialize config and services
	_, err := utils.LoadConfig()
	if err != nil {
		errorLog.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize WebSocket handler
	wsHandler := handlers.NewWebSocketHandler(infoLog, errorLog, debugLog)

	// Set up routes
	http.HandleFunc("/search", wsHandler.HandleSearch)

	// Start server
	serverAddr := ":8080"
	infoLog.Printf("Starting WebSocket server on %s", serverAddr)
	infoLog.Println("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		errorLog.Fatalf("Server failed to start: %v", err)
		os.Exit(1)
	}
}
