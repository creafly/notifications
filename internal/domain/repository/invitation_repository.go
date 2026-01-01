package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.Invitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error)
	GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExpireOld(ctx context.Context) error
}

type invitationRepository struct {
	db *sqlx.DB
}

func NewInvitationRepository(db *sqlx.DB) InvitationRepository {
	return &invitationRepository{db: db}
}

func (r *invitationRepository) Create(ctx context.Context, invitation *entity.Invitation) error {
	query := `
		INSERT INTO invitations (id, tenant_id, tenant_name, inviter_id, inviter_name, invitee_id, email, role_id, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		invitation.ID,
		invitation.TenantID,
		invitation.TenantName,
		invitation.InviterID,
		invitation.InviterName,
		invitation.InviteeID,
		invitation.Email,
		invitation.RoleID,
		invitation.Status,
		invitation.ExpiresAt,
		invitation.CreatedAt,
		invitation.UpdatedAt,
	)
	return err
}

func (r *invitationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
	var invitation entity.Invitation
	query := `SELECT * FROM invitations WHERE id = $1`
	err := r.db.GetContext(ctx, &invitation, query, id)
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *invitationRepository) GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	var invitations []*entity.Invitation
	query := `SELECT * FROM invitations WHERE invitee_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invitations, query, inviteeID)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *invitationRepository) GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	var invitations []*entity.Invitation
	query := `SELECT * FROM invitations WHERE invitee_id = $1 AND status = 'pending' AND expires_at > NOW() ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invitations, query, inviteeID)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *invitationRepository) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error) {
	var invitations []*entity.Invitation
	query := `SELECT * FROM invitations WHERE tenant_id = $1 ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &invitations, query, tenantID)
	if err != nil {
		return nil, err
	}
	return invitations, nil
}

func (r *invitationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.InvitationStatus) error {
	query := `UPDATE invitations SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *invitationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *invitationRepository) ExpireOld(ctx context.Context) error {
	query := `UPDATE invitations SET status = 'expired', updated_at = $1 WHERE status = 'pending' AND expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query, time.Now())
	return err
}
