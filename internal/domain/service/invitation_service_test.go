package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/testutil/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notificationServiceMock struct {
	CreateFunc func(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error)
}

func (m *notificationServiceMock) Create(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return &entity.Notification{ID: uuid.New()}, nil
}

func (m *notificationServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	return nil, nil
}

func (m *notificationServiceMock) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	return nil, nil
}

func (m *notificationServiceMock) GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	return nil, nil
}

func (m *notificationServiceMock) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *notificationServiceMock) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return nil
}

func (m *notificationServiceMock) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}

func (m *notificationServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestInvitationService_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			CreateFunc: func(ctx context.Context, invitation *entity.Invitation) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		input := CreateInvitationInput{
			TenantID:    uuid.New(),
			TenantName:  "Test Tenant",
			InviterID:   uuid.New(),
			InviterName: "Inviter",
			InviteeID:   uuid.New(),
			Email:       "test@example.com",
		}

		result, err := svc.Create(ctx, input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, input.TenantID, result.TenantID)
		assert.Equal(t, input.Email, result.Email)
		assert.Equal(t, entity.InvitationStatusPending, result.Status)
	})

	t.Run("repo error", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			CreateFunc: func(ctx context.Context, invitation *entity.Invitation) error {
				return errors.New("db error")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		input := CreateInvitationInput{
			TenantID:    uuid.New(),
			TenantName:  "Test Tenant",
			InviterID:   uuid.New(),
			InviterName: "Inviter",
			InviteeID:   uuid.New(),
			Email:       "test@example.com",
		}

		_, err := svc.Create(ctx, input)
		assert.Error(t, err)
	})
}

func TestInvitationService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:       invitationID,
			TenantID: uuid.New(),
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		result, err := svc.GetByID(ctx, invitationID)
		require.NoError(t, err)
		assert.Equal(t, invitationID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return nil, errors.New("not found")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		_, err := svc.GetByID(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrInvitationNotFound)
	})
}

func TestInvitationService_Accept(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:        invitationID,
			TenantID:  uuid.New(),
			InviterID: uuid.New(),
			Status:    entity.InvitationStatusPending,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Accept(ctx, invitationID)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return nil, errors.New("not found")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Accept(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrInvitationNotFound)
	})

	t.Run("not pending", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:     invitationID,
			Status: entity.InvitationStatusAccepted,
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Accept(ctx, invitationID)
		assert.ErrorIs(t, err, ErrInvitationNotPending)
	})

	t.Run("expired", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:        invitationID,
			Status:    entity.InvitationStatusPending,
			ExpiresAt: time.Now().Add(-24 * time.Hour),
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Accept(ctx, invitationID)
		assert.ErrorIs(t, err, ErrInvitationExpired)
	})
}

func TestInvitationService_Reject(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:        invitationID,
			TenantID:  uuid.New(),
			InviterID: uuid.New(),
			Status:    entity.InvitationStatusPending,
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
			UpdateStatusFunc: func(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Reject(ctx, invitationID)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return nil, errors.New("not found")
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Reject(ctx, uuid.New())
		assert.ErrorIs(t, err, ErrInvitationNotFound)
	})

	t.Run("not pending", func(t *testing.T) {
		invitationID := uuid.New()
		invitation := &entity.Invitation{
			ID:     invitationID,
			Status: entity.InvitationStatusRejected,
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByIDFunc: func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
				return invitation, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Reject(ctx, invitationID)
		assert.ErrorIs(t, err, ErrInvitationNotPending)
	})
}

func TestInvitationService_Cancel(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitationRepoMock := &mocks.InvitationRepositoryMock{
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		err := svc.Cancel(ctx, uuid.New())
		require.NoError(t, err)
	})
}

func TestInvitationService_GetByInviteeID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		inviteeID := uuid.New()
		invitations := []*entity.Invitation{
			{ID: uuid.New(), InviteeID: inviteeID},
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByInviteeIDFunc: func(ctx context.Context, id uuid.UUID) ([]*entity.Invitation, error) {
				return invitations, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		results, err := svc.GetByInviteeID(ctx, inviteeID)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestInvitationService_GetPendingByInviteeID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		inviteeID := uuid.New()
		invitations := []*entity.Invitation{
			{ID: uuid.New(), InviteeID: inviteeID, Status: entity.InvitationStatusPending},
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetPendingByInviteeIDFunc: func(ctx context.Context, id uuid.UUID) ([]*entity.Invitation, error) {
				return invitations, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		results, err := svc.GetPendingByInviteeID(ctx, inviteeID)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}

func TestInvitationService_GetByTenantID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		tenantID := uuid.New()
		invitations := []*entity.Invitation{
			{ID: uuid.New(), TenantID: tenantID},
		}

		invitationRepoMock := &mocks.InvitationRepositoryMock{
			GetByTenantIDFunc: func(ctx context.Context, id uuid.UUID) ([]*entity.Invitation, error) {
				return invitations, nil
			},
		}
		outboxMock := &mocks.OutboxRepositoryMock{}
		notifSvc := &notificationServiceMock{}

		svc := NewInvitationService(invitationRepoMock, outboxMock, notifSvc)

		results, err := svc.GetByTenantID(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, results, 1)
	})
}
