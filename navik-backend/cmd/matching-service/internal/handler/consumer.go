package handler

import (
	"encoding/json"
	"log"
	"sync"

	"matching-service/internal/model"
	"matching-service/internal/service"

	"github.com/IBM/sarama"
)

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler interface
type ConsumerGroupHandler struct {
	service      service.MatchingService
	activeWorkers sync.WaitGroup
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
	for message := range claim.Messages() {
		select {
		case <-session.Context().Done():
			return nil
		default:
			h.activeWorkers.Add(1)
			go func(msg *sarama.ConsumerMessage) {
				defer h.activeWorkers.Done()
				
				var userLoc model.UserLocation
				if err := json.Unmarshal(msg.Value, &userLoc); err != nil {
					log.Printf("Error unmarshaling message: %v", err)
					session.MarkMessage(msg, "")
					return
				}
				
				if err := h.service.ProcessUserLocation(session.Context(), userLoc); err != nil {
					log.Printf("Error processing user location: %v", err)
				}
				
				session.MarkMessage(msg, "")
			}(message)
		}
	}
	return nil
}
