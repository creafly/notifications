package service

import (
	"context"
	"errors"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/domain/repository"
	"github.com/creafly/notifications/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrPushNotificationNotFound = errors.New("push notification not found")
	ErrPushAlreadySent          = errors.New("push notification already sent")
	ErrInvalidTargetType        = errors.New("invalid target type")
)

type PushNotificationService interface {
	Create(ctx context.Context, input CreatePushInput) (*entity.PushNotification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PushNotification, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.PushNotification, int, error)
	Update(ctx context.Context, id uuid.UUID, input UpdatePushInput) (*entity.PushNotification, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Send(ctx context.Context, id uuid.UUID, userIDs []uuid.UUID) error
	SendImmediate(ctx context.Context, input CreatePushInput, userIDs []uuid.UUID) (*entity.PushNotification, error)
	Cancel(ctx context.Context, id uuid.UUID) error
	GetUserPushNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PushNotification, error)
	MarkAsRead(ctx context.Context, pushID, userID uuid.UUID) error
}

type CreatePushInput struct {
	Title          string
	Message        string
	TargetType     entity.PushTargetType
	TargetTenantID *uuid.UUID
	TargetUserIDs  []uuid.UUID
	Buttons        []entity.PushButton
	ScheduledAt    *time.Time
	CreatedBy      uuid.UUID
}

type UpdatePushInput struct {
	Title          *string
	Message        *string
	TargetType     *entity.PushTargetType
	TargetTenantID *uuid.UUID
	TargetUserIDs  []uuid.UUID
	Buttons        []entity.PushButton
	ScheduledAt    *time.Time
}

type pushNotificationService struct {
	repo            repository.PushNotificationRepository
	notificationSvc NotificationService
	hub             WebSocketHub
}

func NewPushNotificationService(
	repo repository.PushNotificationRepository,
	notificationSvc NotificationService,
	hub WebSocketHub,
) PushNotificationService {
	return &pushNotificationService{
		repo:            repo,
		notificationSvc: notificationSvc,
		hub:             hub,
	}
}

func (s *pushNotificationService) Create(ctx context.Context, input CreatePushInput) (*entity.PushNotification, error) {
	status := entity.PushStatusDraft
	if input.ScheduledAt != nil {
		status = entity.PushStatusScheduled
	}

	push := &entity.PushNotification{
		ID:             utils.GenerateUUID(),
		Title:          input.Title,
		Message:        input.Message,
		TargetType:     input.TargetType,
		TargetTenantID: input.TargetTenantID,
		TargetUserIDs:  input.TargetUserIDs,
		Buttons:        input.Buttons,
		ScheduledAt:    input.ScheduledAt,
		Status:         status,
		CreatedBy:      input.CreatedBy,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.repo.Create(ctx, push); err != nil {
		return nil, err
	}

	return push, nil
}

func (s *pushNotificationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.PushNotification, error) {
	push, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPushNotificationNotFound
	}
	return push, nil
}

func (s *pushNotificationService) GetAll(ctx context.Context, limit, offset int) ([]*entity.PushNotification, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	pushes, err := s.repo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return pushes, total, nil
}

func (s *pushNotificationService) Update(ctx context.Context, id uuid.UUID, input UpdatePushInput) (*entity.PushNotification, error) {
	push, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrPushNotificationNotFound
	}

	if push.Status == entity.PushStatusSent {
		return nil, ErrPushAlreadySent
	}

	if input.Title != nil {
		push.Title = *input.Title
	}
	if input.Message != nil {
		push.Message = *input.Message
	}
	if input.TargetType != nil {
		push.TargetType = *input.TargetType
	}
	if input.TargetTenantID != nil {
		push.TargetTenantID = input.TargetTenantID
	}
	if input.TargetUserIDs != nil {
		push.TargetUserIDs = input.TargetUserIDs
	}
	if input.Buttons != nil {
		push.Buttons = input.Buttons
	}
	if input.ScheduledAt != nil {
		push.ScheduledAt = input.ScheduledAt
		push.Status = entity.PushStatusScheduled
	}

	push.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, push); err != nil {
		return nil, err
	}

	return push, nil
}

func (s *pushNotificationService) Delete(ctx context.Context, id uuid.UUID) error {
	push, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrPushNotificationNotFound
	}

	if push.Status == entity.PushStatusSent {
		return ErrPushAlreadySent
	}

	return s.repo.Delete(ctx, id)
}

func (s *pushNotificationService) Send(ctx context.Context, id uuid.UUID, userIDs []uuid.UUID) error {
	push, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrPushNotificationNotFound
	}

	if push.Status == entity.PushStatusSent {
		return ErrPushAlreadySent
	}

	return s.sendToUsers(ctx, push, userIDs)
}

func (s *pushNotificationService) SendImmediate(ctx context.Context, input CreatePushInput, userIDs []uuid.UUID) (*entity.PushNotification, error) {
	push, err := s.Create(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.sendToUsers(ctx, push, userIDs); err != nil {
		return nil, err
	}

	return push, nil
}

func (s *pushNotificationService) sendToUsers(ctx context.Context, push *entity.PushNotification, userIDs []uuid.UUID) error {
	now := time.Now()

	recipients := make([]*entity.PushNotificationRecipient, len(userIDs))
	for i, userID := range userIDs {
		recipients[i] = &entity.PushNotificationRecipient{
			ID:                 utils.GenerateUUID(),
			PushNotificationID: push.ID,
			UserID:             userID,
			CreatedAt:          now,
		}
	}

	if err := s.repo.CreateRecipients(ctx, recipients); err != nil {
		return err
	}

	dataMap := map[string]interface{}{
		"pushId":  push.ID.String(),
		"buttons": push.Buttons,
	}

	for _, userID := range userIDs {
		if s.notificationSvc != nil {
			_, _ = s.notificationSvc.Create(ctx, CreateNotificationInput{
				UserID:   userID,
				TenantID: push.TargetTenantID,
				Type:     entity.NotificationType("push"),
				Title:    push.Title,
				Message:  push.Message,
				Data:     dataMap,
			})
		}

		_ = s.repo.MarkRecipientDelivered(ctx, push.ID, userID)
	}

	push.Status = entity.PushStatusSent
	push.SentAt = &now
	push.UpdatedAt = now

	return s.repo.Update(ctx, push)
}

func (s *pushNotificationService) Cancel(ctx context.Context, id uuid.UUID) error {
	push, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrPushNotificationNotFound
	}

	if push.Status == entity.PushStatusSent {
		return ErrPushAlreadySent
	}

	push.Status = entity.PushStatusCancelled
	push.UpdatedAt = time.Now()

	return s.repo.Update(ctx, push)
}

func (s *pushNotificationService) GetUserPushNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PushNotification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return s.repo.GetUserPushNotifications(ctx, userID, limit, offset)
}

func (s *pushNotificationService) MarkAsRead(ctx context.Context, pushID, userID uuid.UUID) error {
	return s.repo.MarkRecipientRead(ctx, pushID, userID)
}
