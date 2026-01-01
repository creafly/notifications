package entity

import (
	"time"

	"github.com/google/uuid"
)

type NotificationType string

const (
	NotificationTypeInvitation         NotificationType = "invitation"
	NotificationTypeInvitationAccepted NotificationType = "invitation_accepted"
	NotificationTypeInvitationRejected NotificationType = "invitation_rejected"
	NotificationTypeSystem             NotificationType = "system"
	NotificationTypeSubscription       NotificationType = "subscription"
	NotificationTypeUsage              NotificationType = "usage"
)

type NotificationStatus string

const (
	NotificationStatusUnread   NotificationStatus = "unread"
	NotificationStatusRead     NotificationStatus = "read"
	NotificationStatusArchived NotificationStatus = "archived"
)

type Notification struct {
	ID        uuid.UUID          `db:"id" json:"id"`
	UserID    uuid.UUID          `db:"user_id" json:"userId"`
	TenantID  *uuid.UUID         `db:"tenant_id" json:"tenantId,omitempty"`
	Type      NotificationType   `db:"type" json:"type"`
	Title     string             `db:"title" json:"title"`
	Message   string             `db:"message" json:"message"`
	Data      *string            `db:"data" json:"data,omitempty"`
	Status    NotificationStatus `db:"status" json:"status"`
	ReadAt    *time.Time         `db:"read_at" json:"readAt,omitempty"`
	CreatedAt time.Time          `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `db:"updated_at" json:"updatedAt"`
}

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusExpired  InvitationStatus = "expired"
)

type Invitation struct {
	ID          uuid.UUID        `db:"id" json:"id"`
	TenantID    uuid.UUID        `db:"tenant_id" json:"tenantId"`
	TenantName  string           `db:"tenant_name" json:"tenantName"`
	InviterID   uuid.UUID        `db:"inviter_id" json:"inviterId"`
	InviterName string           `db:"inviter_name" json:"inviterName"`
	InviteeID   uuid.UUID        `db:"invitee_id" json:"inviteeId"`
	Email       string           `db:"email" json:"email"`
	RoleID      *uuid.UUID       `db:"role_id" json:"roleId,omitempty"`
	Status      InvitationStatus `db:"status" json:"status"`
	ExpiresAt   time.Time        `db:"expires_at" json:"expiresAt"`
	CreatedAt   time.Time        `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time        `db:"updated_at" json:"updatedAt"`
}

type OutboxEvent struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	EventType   string     `db:"event_type" json:"eventType"`
	Payload     string     `db:"payload" json:"payload"`
	Status      string     `db:"status" json:"status"`
	RetryCount  int        `db:"retry_count" json:"retryCount"`
	NextRetryAt *time.Time `db:"next_retry_at" json:"nextRetryAt,omitempty"`
	LastErrorAt *time.Time `db:"last_error_at" json:"lastErrorAt,omitempty"`
	ProcessedAt *time.Time `db:"processed_at" json:"processedAt,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
}
