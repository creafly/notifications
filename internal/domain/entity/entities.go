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

type PushTargetType string

const (
	PushTargetAll    PushTargetType = "all"
	PushTargetTenant PushTargetType = "tenant"
	PushTargetUsers  PushTargetType = "users"
)

type PushStatus string

const (
	PushStatusDraft     PushStatus = "draft"
	PushStatusScheduled PushStatus = "scheduled"
	PushStatusSent      PushStatus = "sent"
	PushStatusCancelled PushStatus = "cancelled"
)

type PushButton struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type PushNotification struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	Title          string         `db:"title" json:"title"`
	Message        string         `db:"message" json:"message"`
	TargetType     PushTargetType `db:"target_type" json:"targetType"`
	TargetTenantID *uuid.UUID     `db:"target_tenant_id" json:"targetTenantId,omitempty"`
	TargetUserIDs  []uuid.UUID    `db:"target_user_ids" json:"targetUserIds,omitempty"`
	Buttons        []PushButton   `db:"buttons" json:"buttons,omitempty"`
	ScheduledAt    *time.Time     `db:"scheduled_at" json:"scheduledAt,omitempty"`
	SentAt         *time.Time     `db:"sent_at" json:"sentAt,omitempty"`
	Status         PushStatus     `db:"status" json:"status"`
	CreatedBy      uuid.UUID      `db:"created_by" json:"createdBy"`
	CreatedAt      time.Time      `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time      `db:"updated_at" json:"updatedAt"`
}

type PushNotificationRecipient struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	PushNotificationID uuid.UUID  `db:"push_notification_id" json:"pushNotificationId"`
	UserID             uuid.UUID  `db:"user_id" json:"userId"`
	DeliveredAt        *time.Time `db:"delivered_at" json:"deliveredAt,omitempty"`
	ReadAt             *time.Time `db:"read_at" json:"readAt,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
}
