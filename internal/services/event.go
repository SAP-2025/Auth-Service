package services

import (
	"encoding/json"
	"github.com/SAP-2025/auth-service/internal/config"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
)

type EventService struct {
	publisher message.Publisher
}

func NewEventService(cfg *config.Config) *EventService {
	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:   cfg.Kafka.Brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}, watermill.NewStdLogger(false, false))
	if err != nil {
		panic(err) // Handle properly in prod
	}
	return &EventService{publisher: pub}
}

func (e *EventService) PublishEvent(eventType string, data map[string]interface{}) error {
	payload, _ := json.Marshal(map[string]interface{}{
		"eventType": eventType,
		"version":   "1.0",
		"data":      data,
		"metadata": map[string]string{
			"correlationId": watermill.NewUUID(),
		},
	})
	msg := message.NewMessage(watermill.NewUUID(), payload)
	return e.publisher.Publish("auth_events", msg)
}

func (e *EventService) PublishLoginEvent(userID uint, username, role, loginMethod, provider, ip, ua string) error {
	data := map[string]interface{}{
		"userId":      userID,
		"username":    username,
		"role":        role,
		"loginMethod": loginMethod,
		"provider":    provider,
		"ipAddress":   ip,
		"userAgent":   ua,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}
	return e.PublishEvent("auth.user.login", data)
}

func (e *EventService) PublishLogoutEvent(userID uint, sessionID, reason, ip, ua string) error {
	data := map[string]interface{}{
		"userId":    userID,
		"sessionId": sessionID,
		"reason":    reason,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return e.PublishEvent("auth.user.logout", data)
}

func (e *EventService) PublishTokenRefreshedEvent(userID uint, sessionID, ip, ua string) error {
	data := map[string]interface{}{
		"userId":    userID,
		"sessionId": sessionID,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return e.PublishEvent("auth.token.refreshed", data)
}
