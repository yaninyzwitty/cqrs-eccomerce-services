package messaging

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/apache/pulsar-client-go/pulsar"
)

type MessageProducer interface {
	Publish(ctx context.Context, eventType string, key string, payload []byte) error
}

type PulsarProducer struct {
	producer pulsar.Producer
}

func NewPulsarProducer(producer pulsar.Producer) *PulsarProducer {
	return &PulsarProducer{producer: producer}
}

func (p *PulsarProducer) Publish(ctx context.Context, eventType string, key string, payload []byte) error {
	if p.producer == nil {
		return fmt.Errorf("producer is nil, cannot send messages")
	}

	messageChan := make(chan error, 1)

	p.producer.SendAsync(ctx, &pulsar.ProducerMessage{
		Key:     key,
		Payload: payload,
	}, func(_ pulsar.MessageID, _ *pulsar.ProducerMessage, err error) {
		messageChan <- err
	})

	select {
	case err := <-messageChan:
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("context canceled while publishing message")
	}

	slog.Info("Message sent to Pulsar", "key", key)
	return nil
}
