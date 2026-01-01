package mocks

import (
	"context"
	"time"

	"github.com/creafly/outbox"
	"github.com/google/uuid"
)

type OutboxRepositoryMock struct {
	CreateFunc             func(ctx context.Context, event *outbox.Event) error
	GetPendingFunc         func(ctx context.Context, limit int, maxRetries int) ([]*outbox.Event, error)
	GetPendingForRetryFunc func(ctx context.Context, limit int, maxRetries int) ([]*outbox.Event, error)
	MarkProcessedFunc      func(ctx context.Context, id uuid.UUID) error
	MarkFailedFunc         func(ctx context.Context, id uuid.UUID) error
	IncrementRetryFunc     func(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	CleanupProcessedFunc   func(ctx context.Context, olderThan time.Duration) (int64, error)
	CleanupFailedFunc      func(ctx context.Context, olderThan time.Duration) (int64, error)
}

func (m *OutboxRepositoryMock) Create(ctx context.Context, event *outbox.Event) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return nil
}

func (m *OutboxRepositoryMock) GetPending(ctx context.Context, limit int, maxRetries int) ([]*outbox.Event, error) {
	if m.GetPendingFunc != nil {
		return m.GetPendingFunc(ctx, limit, maxRetries)
	}
	return nil, nil
}

func (m *OutboxRepositoryMock) GetPendingForRetry(ctx context.Context, limit int, maxRetries int) ([]*outbox.Event, error) {
	if m.GetPendingForRetryFunc != nil {
		return m.GetPendingForRetryFunc(ctx, limit, maxRetries)
	}
	return nil, nil
}

func (m *OutboxRepositoryMock) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	if m.MarkProcessedFunc != nil {
		return m.MarkProcessedFunc(ctx, id)
	}
	return nil
}

func (m *OutboxRepositoryMock) MarkFailed(ctx context.Context, id uuid.UUID) error {
	if m.MarkFailedFunc != nil {
		return m.MarkFailedFunc(ctx, id)
	}
	return nil
}

func (m *OutboxRepositoryMock) IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	if m.IncrementRetryFunc != nil {
		return m.IncrementRetryFunc(ctx, id, nextRetryAt)
	}
	return nil
}

func (m *OutboxRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *OutboxRepositoryMock) CleanupProcessed(ctx context.Context, olderThan time.Duration) (int64, error) {
	if m.CleanupProcessedFunc != nil {
		return m.CleanupProcessedFunc(ctx, olderThan)
	}
	return 0, nil
}

func (m *OutboxRepositoryMock) CleanupFailed(ctx context.Context, olderThan time.Duration) (int64, error) {
	if m.CleanupFailedFunc != nil {
		return m.CleanupFailedFunc(ctx, olderThan)
	}
	return 0, nil
}
