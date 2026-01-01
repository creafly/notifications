package mocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
)

type OutboxRepositoryMock struct {
	CreateFunc             func(ctx context.Context, event *entity.OutboxEvent) error
	GetPendingFunc         func(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	GetPendingForRetryFunc func(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	MarkAsProcessedFunc    func(ctx context.Context, id uuid.UUID) error
	MarkAsFailedFunc       func(ctx context.Context, id uuid.UUID) error
	IncrementRetryFunc     func(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error
	DeleteFunc             func(ctx context.Context, id uuid.UUID) error
	CleanupOldFunc         func(ctx context.Context, olderThan time.Duration) error
	CleanupFailedFunc      func(ctx context.Context, olderThan time.Duration) error
}

func (m *OutboxRepositoryMock) Create(ctx context.Context, event *entity.OutboxEvent) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, event)
	}
	return nil
}

func (m *OutboxRepositoryMock) GetPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	return m.GetPendingFunc(ctx, limit)
}

func (m *OutboxRepositoryMock) GetPendingForRetry(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	return m.GetPendingForRetryFunc(ctx, limit)
}

func (m *OutboxRepositoryMock) MarkAsProcessed(ctx context.Context, id uuid.UUID) error {
	return m.MarkAsProcessedFunc(ctx, id)
}

func (m *OutboxRepositoryMock) MarkAsFailed(ctx context.Context, id uuid.UUID) error {
	return m.MarkAsFailedFunc(ctx, id)
}

func (m *OutboxRepositoryMock) IncrementRetry(ctx context.Context, id uuid.UUID, nextRetryAt time.Time) error {
	return m.IncrementRetryFunc(ctx, id, nextRetryAt)
}

func (m *OutboxRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}

func (m *OutboxRepositoryMock) CleanupOld(ctx context.Context, olderThan time.Duration) error {
	return m.CleanupOldFunc(ctx, olderThan)
}

func (m *OutboxRepositoryMock) CleanupFailed(ctx context.Context, olderThan time.Duration) error {
	return m.CleanupFailedFunc(ctx, olderThan)
}
