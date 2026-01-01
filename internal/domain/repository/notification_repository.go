package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/notifications/internal/domain/entity"
	"github.com/jmoiron/sqlx"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *entity.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsRead(ctx context.Context, id uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type notificationRepository struct {
	db *sqlx.DB
}

func NewNotificationRepository(db *sqlx.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	query := `
		INSERT INTO notifications (id, user_id, tenant_id, type, title, message, data, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.ExecContext(ctx, query,
		notification.ID,
		notification.UserID,
		notification.TenantID,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.Data,
		notification.Status,
		notification.CreatedAt,
		notification.UpdatedAt,
	)
	return err
}

func (r *notificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	var notification entity.Notification
	query := `SELECT * FROM notifications WHERE id = $1`
	err := r.db.GetContext(ctx, &notification, query, id)
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	var notifications []*entity.Notification
	query := `SELECT * FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &notifications, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	var notifications []*entity.Notification
	query := `SELECT * FROM notifications WHERE user_id = $1 AND status = 'unread' ORDER BY created_at DESC`
	err := r.db.SelectContext(ctx, &notifications, query, userID)
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND status = 'unread'`
	err := r.db.GetContext(ctx, &count, query, userID)
	return count, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE notifications SET status = 'read', read_at = $1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE notifications SET status = 'read', read_at = $1, updated_at = $1 WHERE user_id = $2 AND status = 'unread'`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	return err
}

func (r *notificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM notifications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
