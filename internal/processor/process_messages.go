package processor

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/yaninyzwitty/cqrs-eccomerce-service/internal/events"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/internal/messaging"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/internal/repository"
)

type ProcessMessage struct {
	producer messaging.MessageProducer
	repo     repository.OutboxRepository
}

func NewProcessMessage(producer messaging.MessageProducer, repo repository.OutboxRepository) *ProcessMessage {
	return &ProcessMessage{producer: producer, repo: repo}
}

func (pm *ProcessMessage) ProcessMessages(ctx context.Context) error {
	bucket := time.Now().Format("2006-01-02")
	messages, err := pm.repo.FetchMessages(ctx, bucket)
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	for _, message := range messages {
		payload, err := pm.handleEvent(message)
		if err != nil {
			slog.Error("Failed to process event", "error", err, "eventType", message.EventType)
			continue
		}

		key := fmt.Sprintf("%s:%v", message.EventType, message.Id)
		if err := pm.producer.Publish(ctx, message.EventType, key, payload); err != nil {
			slog.Error("Failed to send message to Pulsar", "error", err, "messageID", message.Id)
			continue
		}

		if err := pm.repo.DeleteMessage(ctx, message); err != nil {
			slog.Error("Failed to delete message", "error", err, "messageID", message.Id)
		}
	}

	return nil
}

func (pm *ProcessMessage) handleEvent(message repository.OutboxMessage) ([]byte, error) {
	switch message.EventType {
	case "category.created":
		return events.HandleCategoryCreated(message.Payload)
	case "product.created":
		return events.HandleProductCreated(message.Payload)
	default:
		slog.Warn("Unknown event type, skipping", "eventType", message.EventType, "messageID", message.Id)
		return nil, nil
	}
}
