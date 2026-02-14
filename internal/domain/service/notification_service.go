package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/domain/repository"
	"github.com/creafly/notifications/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
)

type NotificationService interface {
	Create(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CreateNotificationInput struct {
	UserID   uuid.UUID
	TenantID *uuid.UUID
	Type     entity.NotificationType
	Title    string
	Message  string
	Data     interface{}
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  WebSocketHub
}

type WebSocketHub interface {
	SendToUser(userID uuid.UUID, message []byte)
}

func NewNotificationService(repo repository.NotificationRepository, hub WebSocketHub) NotificationService {
	return &notificationService{
		repo: repo,
		hub:  hub,
	}
}

func (s *notificationService) Create(ctx context.Context, input CreateNotificationInput) (*entity.Notification, error) {
	var dataJSON *string
	if input.Data != nil {
		data, err := json.Marshal(input.Data)
		if err != nil {
			return nil, err
		}
		dataStr := string(data)
		dataJSON = &dataStr
	}

	notification := &entity.Notification{
		ID:        utils.GenerateUUID(),
		UserID:    input.UserID,
		TenantID:  input.TenantID,
		Type:      input.Type,
		Title:     input.Title,
		Message:   input.Message,
		Data:      dataJSON,
		Status:    entity.NotificationStatusUnread,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		return nil, err
	}

	if s.hub != nil {
		wsMessage, _ := json.Marshal(map[string]interface{}{
			"type":    "notification",
			"payload": notification,
		})
		s.hub.SendToUser(input.UserID, wsMessage)
	}

	return notification, nil
}

func (s *notificationService) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	notification, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotificationNotFound
	}
	return notification, nil
}

func (s *notificationService) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

func (s *notificationService) GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	return s.repo.GetUnreadByUserID(ctx, userID)
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	return s.repo.MarkAsRead(ctx, id)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
