package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/domain/repository"
	"github.com/creafly/outbox"
	"github.com/google/uuid"
)

var (
	ErrInvitationNotFound   = errors.New("invitation not found")
	ErrInvitationExpired    = errors.New("invitation has expired")
	ErrInvitationNotPending = errors.New("invitation is not pending")
)

type InvitationService interface {
	Create(ctx context.Context, input CreateInvitationInput) (*entity.Invitation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error)
	GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error)
	Accept(ctx context.Context, id uuid.UUID) error
	Reject(ctx context.Context, id uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID) error
}

type CreateInvitationInput struct {
	TenantID    uuid.UUID
	TenantName  string
	InviterID   uuid.UUID
	InviterName string
	InviteeID   uuid.UUID
	Email       string
	RoleID      *uuid.UUID
}

type invitationService struct {
	repo                repository.InvitationRepository
	outboxRepo          outbox.Repository
	notificationService NotificationService
}

func NewInvitationService(
	repo repository.InvitationRepository,
	outboxRepo outbox.Repository,
	notificationService NotificationService,
) InvitationService {
	return &invitationService{
		repo:                repo,
		outboxRepo:          outboxRepo,
		notificationService: notificationService,
	}
}

func (s *invitationService) Create(ctx context.Context, input CreateInvitationInput) (*entity.Invitation, error) {
	invitation := &entity.Invitation{
		ID:          uuid.New(),
		TenantID:    input.TenantID,
		TenantName:  input.TenantName,
		InviterID:   input.InviterID,
		InviterName: input.InviterName,
		InviteeID:   input.InviteeID,
		Email:       input.Email,
		RoleID:      input.RoleID,
		Status:      entity.InvitationStatusPending,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	event := outbox.NewEvent("invitations.created", mustMarshal(invitation))
	_ = s.outboxRepo.Create(ctx, event)

	_, _ = s.notificationService.Create(ctx, CreateNotificationInput{
		UserID:   input.InviteeID,
		TenantID: &input.TenantID,
		Type:     entity.NotificationTypeInvitation,
		Title:    "Workspace Invitation",
		Message:  input.InviterName + " invited you to join " + input.TenantName,
		Data: map[string]interface{}{
			"invitationId": invitation.ID,
			"tenantId":     invitation.TenantID,
			"tenantName":   invitation.TenantName,
			"inviterName":  input.InviterName,
		},
	})

	return invitation, nil
}

func (s *invitationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
	invitation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrInvitationNotFound
	}
	return invitation, nil
}

func (s *invitationService) GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	return s.repo.GetByInviteeID(ctx, inviteeID)
}

func (s *invitationService) GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	return s.repo.GetPendingByInviteeID(ctx, inviteeID)
}

func (s *invitationService) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error) {
	return s.repo.GetByTenantID(ctx, tenantID)
}

func (s *invitationService) Accept(ctx context.Context, id uuid.UUID) error {
	invitation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	if invitation.Status != entity.InvitationStatusPending {
		return ErrInvitationNotPending
	}

	if time.Now().After(invitation.ExpiresAt) {
		_ = s.repo.UpdateStatus(ctx, id, entity.InvitationStatusExpired)
		return ErrInvitationExpired
	}

	if err := s.repo.UpdateStatus(ctx, id, entity.InvitationStatusAccepted); err != nil {
		return err
	}

	event := outbox.NewEvent("invitations.accepted", mustMarshal(map[string]any{
		"invitationId": invitation.ID,
		"tenantId":     invitation.TenantID,
		"tenantName":   invitation.TenantName,
		"inviteeId":    invitation.InviteeID,
		"email":        invitation.Email,
		"roleId":       invitation.RoleID,
		"inviterId":    invitation.InviterID,
	}))
	_ = s.outboxRepo.Create(ctx, event)

	_, _ = s.notificationService.Create(ctx, CreateNotificationInput{
		UserID:   invitation.InviterID,
		TenantID: &invitation.TenantID,
		Type:     entity.NotificationTypeInvitationAccepted,
		Title:    "Invitation Accepted",
		Message:  invitation.Email + " accepted your invitation to " + invitation.TenantName,
		Data: map[string]any{
			"invitationId": invitation.ID,
			"tenantId":     invitation.TenantID,
			"tenantName":   invitation.TenantName,
			"userName":     invitation.Email,
		},
	})

	return nil
}

func (s *invitationService) Reject(ctx context.Context, id uuid.UUID) error {
	invitation, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	if invitation.Status != entity.InvitationStatusPending {
		return ErrInvitationNotPending
	}

	if err := s.repo.UpdateStatus(ctx, id, entity.InvitationStatusRejected); err != nil {
		return err
	}

	event := outbox.NewEvent("invitations.rejected", mustMarshal(invitation))
	_ = s.outboxRepo.Create(ctx, event)

	_, _ = s.notificationService.Create(ctx, CreateNotificationInput{
		UserID:   invitation.InviterID,
		TenantID: &invitation.TenantID,
		Type:     entity.NotificationTypeInvitationRejected,
		Title:    "Invitation Rejected",
		Message:  invitation.Email + " declined your invitation to " + invitation.TenantName,
		Data: map[string]interface{}{
			"invitationId": invitation.ID,
			"tenantId":     invitation.TenantID,
			"tenantName":   invitation.TenantName,
			"userName":     invitation.Email,
		},
	})

	return nil
}

func (s *invitationService) Cancel(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func mustMarshal(v interface{}) string {
	data, _ := json.Marshal(v)
	return string(data)
}
