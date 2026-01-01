package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/notifications/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

const (
	MaxRetryCount = 10
)

type OutboxRepository interface {
	Create(ctx context.Context, event *entity.OutboxEvent) error
	GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	GetPendingForRetry(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	MarkAsFailed(ctx context.Context, id uuid.UUID) error
	IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
	CleanupOld(ctx context.Context, olderThan time.Duration) error
	CleanupFailed(ctx context.Context, olderThan time.Duration) error
}

type outboxRepository struct {
	db *sqlx.DB
}

func NewOutboxRepository(db *sqlx.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) Create(ctx context.Context, event *entity.OutboxEvent) error {
	query := `
		INSERT INTO outbox_events (id, event_type, payload, status, retry_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.ExecContext(ctx, query,
		event.ID,
		event.EventType,
		event.Payload,
		event.Status,
		event.RetryCount,
		event.CreatedAt,
	)
	return err
}

func (r *outboxRepository) GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	var events []*entity.OutboxEvent
	query := `
		SELECT * FROM outbox_events 
		WHERE status = 'pending' 
		  AND retry_count < $1
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC 
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &events, query, MaxRetryCount, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *outboxRepository) GetPendingForRetry(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	var events []*entity.OutboxEvent
	query := `
		SELECT * FROM outbox_events 
		WHERE status = 'pending' 
		  AND retry_count > 0
		  AND retry_count < $1
		  AND next_retry_at <= NOW()
		ORDER BY next_retry_at ASC 
		LIMIT $2
	`
	err := r.db.SelectContext(ctx, &events, query, MaxRetryCount, limit)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func (r *outboxRepository) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = 'sent', processed_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *outboxRepository) MarkAsFailed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = 'failed' WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *outboxRepository) IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	query := `
		UPDATE outbox_events 
		SET retry_count = retry_count + 1, 
		    next_retry_at = $1,
		    last_error_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, nextRetryAt, id)
	return err
}

func (r *outboxRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM outbox_events WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *outboxRepository) CleanupOld(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM outbox_events WHERE status = 'sent' AND processed_at < $1`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}

func (r *outboxRepository) CleanupFailed(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM outbox_events WHERE status = 'failed' AND created_at < $1`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}
