package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PushNotificationRepository interface {
	Create(ctx context.Context, push *entity.PushNotification) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PushNotification, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entity.PushNotification, error)
	GetByStatus(ctx context.Context, status entity.PushStatus, limit, offset int) ([]*entity.PushNotification, error)
	GetScheduledToSend(ctx context.Context) ([]*entity.PushNotification, error)
	Update(ctx context.Context, push *entity.PushNotification) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)

	CreateRecipients(ctx context.Context, recipients []*entity.PushNotificationRecipient) error
	GetRecipientsByPushID(ctx context.Context, pushID uuid.UUID) ([]*entity.PushNotificationRecipient, error)
	GetUserPushNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PushNotification, error)
	MarkRecipientDelivered(ctx context.Context, pushID, userID uuid.UUID) error
	MarkRecipientRead(ctx context.Context, pushID, userID uuid.UUID) error
}

type pushNotificationRepository struct {
	db *sqlx.DB
}

func NewPushNotificationRepository(db *sqlx.DB) PushNotificationRepository {
	return &pushNotificationRepository{db: db}
}

type pushNotificationRow struct {
	ID             uuid.UUID      `db:"id"`
	Title          string         `db:"title"`
	Message        string         `db:"message"`
	TargetType     string         `db:"target_type"`
	TargetTenantID *uuid.UUID     `db:"target_tenant_id"`
	TargetUserIDs  pq.StringArray `db:"target_user_ids"`
	Buttons        sql.NullString `db:"buttons"`
	ScheduledAt    *time.Time     `db:"scheduled_at"`
	SentAt         *time.Time     `db:"sent_at"`
	Status         string         `db:"status"`
	CreatedBy      uuid.UUID      `db:"created_by"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

func (row *pushNotificationRow) toEntity() *entity.PushNotification {
	push := &entity.PushNotification{
		ID:             row.ID,
		Title:          row.Title,
		Message:        row.Message,
		TargetType:     entity.PushTargetType(row.TargetType),
		TargetTenantID: row.TargetTenantID,
		ScheduledAt:    row.ScheduledAt,
		SentAt:         row.SentAt,
		Status:         entity.PushStatus(row.Status),
		CreatedBy:      row.CreatedBy,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}

	if len(row.TargetUserIDs) > 0 {
		push.TargetUserIDs = make([]uuid.UUID, 0, len(row.TargetUserIDs))
		for _, idStr := range row.TargetUserIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				push.TargetUserIDs = append(push.TargetUserIDs, id)
			}
		}
	}

	if row.Buttons.Valid && row.Buttons.String != "" {
		var buttons []entity.PushButton
		if err := json.Unmarshal([]byte(row.Buttons.String), &buttons); err == nil {
			push.Buttons = buttons
		}
	}

	return push
}

func (r *pushNotificationRepository) Create(ctx context.Context, push *entity.PushNotification) error {
	var buttonsJSON interface{}
	if len(push.Buttons) > 0 {
		data, err := json.Marshal(push.Buttons)
		if err != nil {
			return err
		}
		buttonsJSON = data
	}

	var targetUserIDs interface{}
	if len(push.TargetUserIDs) > 0 {
		ids := make([]string, len(push.TargetUserIDs))
		for i, id := range push.TargetUserIDs {
			ids[i] = id.String()
		}
		targetUserIDs = pq.Array(ids)
	}

	query := `
		INSERT INTO push_notifications (id, title, message, target_type, target_tenant_id, target_user_ids, buttons, scheduled_at, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.ExecContext(ctx, query,
		push.ID,
		push.Title,
		push.Message,
		push.TargetType,
		push.TargetTenantID,
		targetUserIDs,
		buttonsJSON,
		push.ScheduledAt,
		push.Status,
		push.CreatedBy,
		push.CreatedAt,
		push.UpdatedAt,
	)
	return err
}

func (r *pushNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.PushNotification, error) {
	var row pushNotificationRow
	query := `SELECT * FROM push_notifications WHERE id = $1`
	err := r.db.GetContext(ctx, &row, query, id)
	if err != nil {
		return nil, err
	}
	return row.toEntity(), nil
}

func (r *pushNotificationRepository) GetAll(ctx context.Context, limit, offset int) ([]*entity.PushNotification, error) {
	var rows []pushNotificationRow
	query := `SELECT * FROM push_notifications ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	err := r.db.SelectContext(ctx, &rows, query, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.PushNotification, len(rows))
	for i, row := range rows {
		result[i] = row.toEntity()
	}
	return result, nil
}

func (r *pushNotificationRepository) GetByStatus(ctx context.Context, status entity.PushStatus, limit, offset int) ([]*entity.PushNotification, error) {
	var rows []pushNotificationRow
	query := `SELECT * FROM push_notifications WHERE status = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	err := r.db.SelectContext(ctx, &rows, query, status, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.PushNotification, len(rows))
	for i, row := range rows {
		result[i] = row.toEntity()
	}
	return result, nil
}

func (r *pushNotificationRepository) GetScheduledToSend(ctx context.Context) ([]*entity.PushNotification, error) {
	var rows []pushNotificationRow
	query := `SELECT * FROM push_notifications WHERE status = 'scheduled' AND scheduled_at <= NOW() ORDER BY scheduled_at ASC`
	err := r.db.SelectContext(ctx, &rows, query)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.PushNotification, len(rows))
	for i, row := range rows {
		result[i] = row.toEntity()
	}
	return result, nil
}

func (r *pushNotificationRepository) Update(ctx context.Context, push *entity.PushNotification) error {
	var buttonsJSON interface{}
	if len(push.Buttons) > 0 {
		data, err := json.Marshal(push.Buttons)
		if err != nil {
			return err
		}
		buttonsJSON = data
	}

	var targetUserIDs interface{}
	if len(push.TargetUserIDs) > 0 {
		ids := make([]string, len(push.TargetUserIDs))
		for i, id := range push.TargetUserIDs {
			ids[i] = id.String()
		}
		targetUserIDs = pq.Array(ids)
	}

	query := `
		UPDATE push_notifications 
		SET title = $1, message = $2, target_type = $3, target_tenant_id = $4, target_user_ids = $5, 
		    buttons = $6, scheduled_at = $7, sent_at = $8, status = $9, updated_at = $10
		WHERE id = $11
	`
	_, err := r.db.ExecContext(ctx, query,
		push.Title,
		push.Message,
		push.TargetType,
		push.TargetTenantID,
		targetUserIDs,
		buttonsJSON,
		push.ScheduledAt,
		push.SentAt,
		push.Status,
		time.Now(),
		push.ID,
	)
	return err
}

func (r *pushNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM push_notifications WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *pushNotificationRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM push_notifications`
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}

func (r *pushNotificationRepository) CreateRecipients(ctx context.Context, recipients []*entity.PushNotificationRecipient) error {
	if len(recipients) == 0 {
		return nil
	}

	query := `
		INSERT INTO push_notification_recipients (id, push_notification_id, user_id, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (push_notification_id, user_id) DO NOTHING
	`

	for _, r2 := range recipients {
		_, err := r.db.ExecContext(ctx, query, r2.ID, r2.PushNotificationID, r2.UserID, r2.CreatedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *pushNotificationRepository) GetRecipientsByPushID(ctx context.Context, pushID uuid.UUID) ([]*entity.PushNotificationRecipient, error) {
	var recipients []*entity.PushNotificationRecipient
	query := `SELECT * FROM push_notification_recipients WHERE push_notification_id = $1`
	err := r.db.SelectContext(ctx, &recipients, query, pushID)
	if err != nil {
		return nil, err
	}
	return recipients, nil
}

func (r *pushNotificationRepository) GetUserPushNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.PushNotification, error) {
	var rows []pushNotificationRow
	query := `
		SELECT pn.* FROM push_notifications pn
		INNER JOIN push_notification_recipients pnr ON pn.id = pnr.push_notification_id
		WHERE pnr.user_id = $1 AND pn.status = 'sent'
		ORDER BY pn.sent_at DESC
		LIMIT $2 OFFSET $3
	`
	err := r.db.SelectContext(ctx, &rows, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]*entity.PushNotification, len(rows))
	for i, row := range rows {
		result[i] = row.toEntity()
	}
	return result, nil
}

func (r *pushNotificationRepository) MarkRecipientDelivered(ctx context.Context, pushID, userID uuid.UUID) error {
	query := `UPDATE push_notification_recipients SET delivered_at = $1 WHERE push_notification_id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), pushID, userID)
	return err
}

func (r *pushNotificationRepository) MarkRecipientRead(ctx context.Context, pushID, userID uuid.UUID) error {
	query := `UPDATE push_notification_recipients SET read_at = $1 WHERE push_notification_id = $2 AND user_id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), pushID, userID)
	return err
}
