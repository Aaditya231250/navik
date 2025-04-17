package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"matching-service/internal/config"
	"matching-service/internal/handler"
	"matching-service/internal/service"
	"matching-service/pkg/kafka"

	"github.com/go-redis/redis/v8"
)

func main() {
	configFile := flag.String("config", "config.json", "Path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.TopicFormat)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "redis:6379", // Use environment variable in production
		DB:       0,            // use default DB
		Password: "",
	})
	defer redisClient.Close()

	locationService := service.NewLocationService(nil, producer)
	locationHandler := handler.NewLocationHandler(locationService)
	wsHandler := handler.NewWebSocketHandler(redisClient)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsHandler.HandleWebSocket)
	locationHandler.SetupRoutes(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
