package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
)

type NotificationRepositoryMock struct {
	CreateFunc            func(ctx context.Context, notification *entity.Notification) error
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	GetByUserIDFunc       func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadByUserIDFunc func(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error)
	GetUnreadCountFunc    func(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsReadFunc        func(ctx context.Context, id uuid.UUID) error
	MarkAllAsReadFunc     func(ctx context.Context, userID uuid.UUID) error
	DeleteFunc            func(ctx context.Context, id uuid.UUID) error
}

func (m *NotificationRepositoryMock) Create(ctx context.Context, notification *entity.Notification) error {
	return m.CreateFunc(ctx, notification)
}

func (m *NotificationRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *NotificationRepositoryMock) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	return m.GetByUserIDFunc(ctx, userID, limit, offset)
}

func (m *NotificationRepositoryMock) GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	return m.GetUnreadByUserIDFunc(ctx, userID)
}

func (m *NotificationRepositoryMock) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return m.GetUnreadCountFunc(ctx, userID)
}

func (m *NotificationRepositoryMock) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return m.MarkAsReadFunc(ctx, id)
}

func (m *NotificationRepositoryMock) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return m.MarkAllAsReadFunc(ctx, userID)
}

func (m *NotificationRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}
