package handler

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"matching-service/internal/model"
	"matching-service/internal/service"

	"github.com/IBM/sarama"
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler interface
type ConsumerGroupHandler struct {
	service       service.MatchingService
	activeWorkers sync.WaitGroup
	processingMap sync.Map // To track messages being processed
}

// NewConsumerGroupHandler creates a new Kafka consumer handler
func NewConsumerGroupHandler(service service.MatchingService) *ConsumerGroupHandler {
	return &ConsumerGroupHandler{
		service: service,
	}
}

// Setup is called when the consumer group session is starting
func (h *ConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session started: %s", session.MemberID())
	return nil
}

// Cleanup is called when the consumer group session is ending
func (h *ConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Printf("Consumer group session ending: %s", session.MemberID())
	h.activeWorkers.Wait()
	return nil
}

// ConsumeClaim processes messages from a partition claim
func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.Printf("Starting to consume from topic: %s, partition: %d, initial offset: %d",
		claim.Topic(), claim.Partition(), claim.InitialOffset())

	// Process messages from the claim
	for message := range claim.Messages() {
		// Check if context is done
		select {
		case <-session.Context().Done():
			return nil
		default:
			// Generate a unique message ID for tracking
			msgID := generateMessageID(message)

			// Check if already processing this message (avoid duplicates in case of rebalance)
			if _, exists := h.processingMap.LoadOrStore(msgID, true); exists {
				log.Printf("Skipping duplicate message: %s", msgID)
				session.MarkMessage(message, "")
				continue
			}

			// Process the message
			h.activeWorkers.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer h.activeWorkers.Done()
				defer h.processingMap.Delete(msgID)

				startTime := time.Now()

				var userLoc model.UserLocation
				if err := json.Unmarshal(msg.Value, &userLoc); err != nil {
					log.Printf("Error unmarshaling message: %v, content: %s",
						err, string(msg.Value))
					session.MarkMessage(msg, "")
					return
				}

				log.Printf("Processing message: topic=%s, partition=%d, offset=%d, key=%s, user=%s",
					msg.Topic, msg.Partition, msg.Offset, string(msg.Key), userLoc.UserID)
				
				if time.Now().Unix() - userLoc.Timestamp > 300 { 
					log.Printf("Skipping stale location update for user %s (%.2f minutes old)",
						userLoc.UserID, float64(time.Now().Unix()-userLoc.Timestamp)/60)
					session.MarkMessage(msg, "")
					return
				}	

				if err := h.service.ProcessUserLocation(session.Context(), userLoc); err != nil {
					log.Printf("Error processing user location: %v", err)
					// You can implement retry logic here if needed
					// For now, we'll mark it as processed to avoid getting stuck
				}

				// Mark message as processed
				session.MarkMessage(msg, "")

				elapsed := time.Since(startTime)
				log.Printf("Finished processing message in %v: topic=%s, partition=%d, offset=%d",
					elapsed, msg.Topic, msg.Partition, msg.Offset)

			}(message)
		}
	}

	log.Printf("Finished consuming from topic: %s, partition: %d",
		claim.Topic(), claim.Partition())
	return nil
}

// generateMessageID creates a unique ID for a message based on topic, partition, and offset
func generateMessageID(msg *sarama.ConsumerMessage) string {
	return msg.Topic + "-" + string(msg.Key)
}
