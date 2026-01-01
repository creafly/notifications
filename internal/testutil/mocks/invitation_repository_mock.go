package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/hexaend/notifications/internal/domain/entity"
)

type InvitationRepositoryMock struct {
	CreateFunc                func(ctx context.Context, invitation *entity.Invitation) error
	GetByIDFunc               func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error)
	GetByInviteeIDFunc        func(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetPendingByInviteeIDFunc func(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetByTenantIDFunc         func(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error)
	UpdateStatusFunc          func(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error
	DeleteFunc                func(ctx context.Context, id uuid.UUID) error
	ExpireOldFunc             func(ctx context.Context) error
}

func (m *InvitationRepositoryMock) Create(ctx context.Context, invitation *entity.Invitation) error {
	return m.CreateFunc(ctx, invitation)
}

func (m *InvitationRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
	return m.GetByIDFunc(ctx, id)
}

func (m *InvitationRepositoryMock) GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	return m.GetByInviteeIDFunc(ctx, inviteeID)
}

func (m *InvitationRepositoryMock) GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	return m.GetPendingByInviteeIDFunc(ctx, inviteeID)
}

func (m *InvitationRepositoryMock) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error) {
	return m.GetByTenantIDFunc(ctx, tenantID)
}

func (m *InvitationRepositoryMock) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error {
	return m.UpdateStatusFunc(ctx, id, status)
}

func (m *InvitationRepositoryMock) Delete(ctx context.Context, id uuid.UUID) error {
	return m.DeleteFunc(ctx, id)
}

func (m *InvitationRepositoryMock) ExpireOld(ctx context.Context) error {
	return m.ExpireOldFunc(ctx)
}
