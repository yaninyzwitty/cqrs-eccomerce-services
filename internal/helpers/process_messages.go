package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/gocql/gocql"
	"github.com/yaninyzwitty/cqrs-eccomerce-service/pb"
)

type ProcessMessage struct {
	producer pulsar.Producer
	session  *gocql.Session
}

type OutboxMessage struct {
	Id        gocql.UUID
	Bucket    string
	EventType EventType
	Payload   string
}

type EventType string

const (
	ProductCreated  EventType = "product.created"
	CategoryCreated EventType = "category.created"
)

var getProductsFromOutboxQuery = `
	SELECT id, bucket, payload, event_type 
	FROM products_keyspace_v3.outbox 
	WHERE bucket = ? 
	ORDER BY id ASC;
`

func NewProcessMessage(producer pulsar.Producer, session *gocql.Session) *ProcessMessage {
	return &ProcessMessage{
		producer: producer,
		session:  session,
	}
}

func (pm *ProcessMessage) ProcessMessages(ctx context.Context) error {
	bucket := getCurrentBucket()

	messages, err := pm.FetchMessages(ctx, bucket)
	if err != nil {
		return fmt.Errorf("error fetching messages: %w", err)
	}

	for _, message := range messages {
		if err := pm.sendToPulsar(ctx, message); err != nil {
			slog.Error("Failed to send message to Pulsar", "error", err, "messageID", message.Id)
			continue
		}
	}
	return nil
}

func getCurrentBucket() string {
	return time.Now().Format("2006-01-02")
}

func (pm *ProcessMessage) FetchMessages(ctx context.Context, bucket string) ([]OutboxMessage, error) {
	var outboxMessages []OutboxMessage
	iter := pm.session.Query(getProductsFromOutboxQuery, bucket).WithContext(ctx).Iter()
	defer iter.Close()

	for {
		var outboxMessage OutboxMessage
		if !iter.Scan(&outboxMessage.Id, &outboxMessage.Bucket, &outboxMessage.Payload, &outboxMessage.EventType) {
			break
		}
		outboxMessages = append(outboxMessages, outboxMessage)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return outboxMessages, nil
}

func (pm *ProcessMessage) sendToPulsar(ctx context.Context, message OutboxMessage) error {
	slog.Info("Sending message to Pulsar", "messageID", message.Id)

	switch message.EventType {
	case CategoryCreated:
		var category pb.Category
		if err := json.Unmarshal([]byte(message.Payload), &category); err != nil {
			return fmt.Errorf("error unmarshalling category: %w", err)
		}
		category.EventType = string(message.EventType)
		payload, err := json.Marshal(&category)
		if err != nil {
			return fmt.Errorf("error marshalling category: %w", err)
		}
		return pm.publishMessage(ctx, message.EventType, message, payload)

	case ProductCreated:
		var product pb.Product
		if err := json.Unmarshal([]byte(message.Payload), &product); err != nil {
			return fmt.Errorf("error unmarshalling product: %w", err)
		}
		product.EventType = string(message.EventType)
		payload, err := json.Marshal(&product)
		if err != nil {
			return fmt.Errorf("error marshalling product: %w", err)
		}
		return pm.publishMessage(ctx, message.EventType, message, payload)

	default:
		slog.Warn("Unknown event type, skipping", "eventType", message.EventType, "messageID", message.Id)
		return nil
	}
}
func (pm *ProcessMessage) publishMessage(ctx context.Context, eventType EventType, message OutboxMessage, payload []byte) error {
	if pm.producer == nil {
		return fmt.Errorf("producer is nil, cannot send messages")
	}

	messageChan := make(chan error, 1)

	pm.producer.SendAsync(ctx, &pulsar.ProducerMessage{
		Key:     fmt.Sprintf("%s:%v", eventType, message.Id),
		Payload: payload,
	}, func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
		select {
		case messageChan <- err:
		default:
		}
	})

	var publishErr error
	select {
	case err := <-messageChan:
		publishErr = err
	case <-ctx.Done():
		publishErr = fmt.Errorf("context canceled while publishing message")
	}

	close(messageChan)

	if publishErr != nil {
		return fmt.Errorf("failed to publish message: %w", publishErr)
	}

	slog.Info("Message sent to Pulsar", "messageID", message.Id, "eventType", eventType)
	return pm.deleteMessage(ctx, message)
}

func (pm *ProcessMessage) deleteMessage(ctx context.Context, message OutboxMessage) error {
	if pm.session == nil {
		return fmt.Errorf("session is nil, cannot delete messages")
	}

	slog.Info("Deleting message", "messageID", message.Id)

	return pm.session.Query(`
		DELETE FROM products_keyspace_v3.outbox WHERE bucket = ? AND id = ?`,
		message.Bucket, message.Id,
	).WithContext(ctx).Exec()
}
