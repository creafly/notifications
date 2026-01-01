package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type wsHubMock struct{}

func (h *wsHubMock) SendToUser(userID uuid.UUID, message []byte) {}

func TestNotificationService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			CreateFunc: func(ctx context.Context, notification *entity.Notification) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		input := CreateNotificationInput{
			UserID:  uuid.New(),
			Type:    entity.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}

		result, err := svc.Create(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, input.UserID, result.UserID)
		assert.Equal(t, input.Title, result.Title)
		assert.Equal(t, entity.NotificationStatusUnread, result.Status)
	})

	t.Run("with websocket hub", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			CreateFunc: func(ctx context.Context, notification *entity.Notification) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		hub := &wsHubMock{}

		svc := NewNotificationService(repoMock, outboxMock, hub)

		input := CreateNotificationInput{
			UserID:  uuid.New(),
			Type:    entity.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}

		result, err := svc.Create(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("repo error", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			CreateFunc: func(ctx context.Context, notification *entity.Notification) error {
				return errors.New("db error")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		input := CreateNotificationInput{
			UserID:  uuid.New(),
			Type:    entity.NotificationTypeSystem,
			Title:   "Test",
			Message: "Test message",
		}

		_, err := svc.Create(ctx, input)
		assert.Error(t, err)
	})
}

func TestNotificationService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		notificationID := uuid.New()
		notification := &entity.Notification{
			ID:     notificationID,
			UserID: uuid.New(),
			Title:  "Test",
		}

		repoMock := &mocks.NotificationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
				return notification, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		result, err := svc.GetByID(ctx, notificationID)
		require.NoError(t, err)
		assert.Equal(t, notificationID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
				return nil, errors.New("not found")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		_, err := svc.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrNotificationNotFound)
	})
}

func TestNotificationService_GetByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		notifications := []*entity.Notification{
			{ID: uuid.New(), UserID: userID},
			{ID: uuid.New(), UserID: userID},
		}

		repoMock := &mocks.NotificationRepositoryMock{
			GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
				return notifications, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		results, err := svc.GetByUserID(ctx, userID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("default limit", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
				assert.Equal(t, 20, limit)
				return nil, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		_, err := svc.GetByUserID(ctx, uuid.New(), 0, 0)
		require.NoError(t, err)
	})

	t.Run("max limit", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
				assert.Equal(t, 100, limit)
				return nil, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		_, err := svc.GetByUserID(ctx, uuid.New(), 500, 0)
		require.NoError(t, err)
	})
}

func TestNotificationService_GetUnreadByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := uuid.New()
		notifications := []*entity.Notification{
			{ID: uuid.New(), UserID: userID, Status: entity.NotificationStatusUnread},
		}

		repoMock := &mocks.NotificationRepositoryMock{
			GetUnreadByUserIDFunc: func(ctx context.Context, uid uuid.UUID) ([]*entity.Notification, error) {
				return notifications, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		results, err := svc.GetUnreadByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestNotificationService_GetUnreadCount(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			GetUnreadCountFunc: func(ctx context.Context, userID uuid.UUID) (int, error) {
				return 5, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		count, err := svc.GetUnreadCount(ctx, uuid.New())
		require.NoError(t, err)
		assert.Equal(t, 5, count)
	})
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		err := svc.MarkAsRead(ctx, uuid.New())
		require.NoError(t, err)
	})
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			MarkAllAsReadFunc: func(ctx context.Context, userID uuid.UUID) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		err := svc.MarkAllAsRead(ctx, uuid.New())
		require.NoError(t, err)
	})
}

func TestNotificationService_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repoMock := &mocks.NotificationRepositoryMock{
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}

		svc := NewNotificationService(repoMock, outboxMock, nil)

		err := svc.Delete(ctx, uuid.New())
		require.NoError(t, err)
	})
}
