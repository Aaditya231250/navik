// internal/handler/websocket.go
package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	redisClient *redis.Client
	upgrader    websocket.Upgrader
}

func NewWebSocketHandler(redisClient *redis.Client) *WebSocketHandler {
	return &WebSocketHandler{
		redisClient: redisClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

// func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
// 	conn, err := h.upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Printf("WebSocket upgrade failed: %v", err)
// 		return
// 	}
// 	defer conn.Close()

// 	// Get userID from query param or cookie
// 	userID := r.URL.Query().Get("user_id")
// 	if userID == "" {
// 		conn.WriteMessage(websocket.TextMessage, []byte("user_id required"))
// 		return
// 	}

// 	ctx := context.Background()
// 	pubsub := h.redisClient.Subscribe(ctx, "user:"+userID)
// 	defer pubsub.Close()

// 	for msg := range pubsub.Channel() {
// 		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
// 			log.Printf("WebSocket write error: %v", err)
// 			return
// 		}
// 	}
// }

//	func (h *WebSocketHandler) SendMatchResults(userID string, data []byte) error {
//		ctx := context.Background()
//		return h.redisClient.Publish(ctx, "user:"+userID, data).Err()
//	}
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Get userID from query param or cookie
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("user_id required"))
		return
	}

	// Set up Redis subscription with better logging
	ctx := context.Background()
	channelName := "user:" + userID
	log.Printf("Subscribing to Redis channel: %s", channelName)

	pubsub := h.redisClient.Subscribe(ctx, channelName)
	defer pubsub.Close()

	// // Send confirmation message to client
	// conn.WriteMessage(websocket.TextMessage, []byte(`{"status":"CONNECTED","message":"Subscribed to updates"}`))

	// Process incoming messages from Redis
	ch := pubsub.Channel()
	log.Printf("Starting Redis message loop for user: %s", userID)

	for msg := range ch {
		log.Printf("Received message from Redis channel %s: %s", channelName, msg.Payload)
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
		log.Printf("Successfully sent message to client: %s", userID)
	}

	log.Printf("Redis subscription channel closed for user: %s", userID)
}

func (h *WebSocketHandler) SendMatchResults(userID string, data []byte) error {
	ctx := context.Background()
	channelName := "user:" + userID
	log.Printf("Publishing match results to channel %s: %s", channelName, string(data))

	result := h.redisClient.Publish(ctx, channelName, data)
	if err := result.Err(); err != nil {
		log.Printf("Failed to publish to Redis: %v", err)
		return err
	}

	recipients, err := result.Result()
	if err != nil {
		log.Printf("Error getting publish result: %v", err)
		return err
	}

	log.Printf("Message published to %d subscribers", recipients)
	if recipients == 0 {
		log.Printf("Warning: No subscribers for channel %s", channelName)
	}

	return nil
}
