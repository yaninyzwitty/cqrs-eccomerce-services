package repository

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/gocql/gocql"
)

type OutboxMessage struct {
	Id        gocql.UUID
	Bucket    string
	EventType string
	Payload   string
}

type OutboxRepository interface {
	FetchMessages(ctx context.Context, bucket string) ([]OutboxMessage, error)
	DeleteMessage(ctx context.Context, message OutboxMessage) error
}

type CassandraOutboxRepository struct {
	session *gocql.Session
}

func NewCassandraOutboxRepository(session *gocql.Session) *CassandraOutboxRepository {
	return &CassandraOutboxRepository{session: session}
}

func (r *CassandraOutboxRepository) FetchMessages(ctx context.Context, bucket string) ([]OutboxMessage, error) {
	query := `SELECT id, bucket, payload, event_type FROM products_keyspace_v3.outbox WHERE bucket = ? ORDER BY id ASC;`
	iter := r.session.Query(query, bucket).WithContext(ctx).Iter()
	defer iter.Close()

	var messages []OutboxMessage
	for {
		var msg OutboxMessage
		if !iter.Scan(&msg.Id, &msg.Bucket, &msg.Payload, &msg.EventType) {
			break
		}
		messages = append(messages, msg)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *CassandraOutboxRepository) DeleteMessage(ctx context.Context, message OutboxMessage) error {
	query := `DELETE FROM products_keyspace_v3.outbox WHERE bucket = ? AND id = ?`
	err := r.session.Query(query, message.Bucket, message.Id).WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	slog.Info("Message deleted", "messageID", message.Id)
	return nil
}
