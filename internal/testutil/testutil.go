package testutil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/jmoiron/sqlx"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type TestDB struct {
	DB *sqlx.DB
}

func SetupTestDB(t *testing.T) *TestDB {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5440/notifications_test?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	runMigrations(t, db)
	cleanupTables(t, db)

	return &TestDB{DB: db}
}

func (tdb *TestDB) Cleanup(t *testing.T) {
	cleanupTables(t, tdb.DB)
	tdb.DB.Close()
}

func runMigrations(t *testing.T, db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		t.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../../migrations",
		"postgres", driver)
	if err != nil {
		t.Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %v", err)
	}
}

func cleanupTables(t *testing.T, db *sqlx.DB) {
	tables := []string{"outbox_events", "invitations", "notifications"}
	for _, table := range tables {
		_, err := db.Exec("DELETE FROM " + table)
		if err != nil {
			t.Logf("Warning: failed to clean table %s: %v", table, err)
		}
	}
}

func NewTestNotification() *entity.Notification {
	now := time.Now()
	return &entity.Notification{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		TenantID:  nil,
		Type:      entity.NotificationTypeSystem,
		Title:     "Test Notification",
		Message:   "This is a test notification message",
		Data:      nil,
		Status:    entity.NotificationStatusUnread,
		ReadAt:    nil,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewTestNotificationWithUser(userID uuid.UUID) *entity.Notification {
	n := NewTestNotification()
	n.UserID = userID
	return n
}

func NewTestNotificationWithTenant(userID, tenantID uuid.UUID) *entity.Notification {
	n := NewTestNotification()
	n.UserID = userID
	n.TenantID = &tenantID
	return n
}

func NewTestInvitation() *entity.Invitation {
	now := time.Now()
	return &entity.Invitation{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		TenantName:  "Test Tenant",
		InviterID:   uuid.New(),
		InviterName: "Test Inviter",
		InviteeID:   uuid.New(),
		Email:       fmt.Sprintf("invitee-%s@example.com", uuid.New().String()[:8]),
		RoleID:      nil,
		Status:      entity.InvitationStatusPending,
		ExpiresAt:   now.Add(7 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func NewTestInvitationWithInvitee(inviteeID uuid.UUID) *entity.Invitation {
	inv := NewTestInvitation()
	inv.InviteeID = inviteeID
	return inv
}

func NewTestInvitationWithTenant(tenantID uuid.UUID) *entity.Invitation {
	inv := NewTestInvitation()
	inv.TenantID = tenantID
	return inv
}

func NewTestInvitationExpired() *entity.Invitation {
	inv := NewTestInvitation()
	inv.ExpiresAt = time.Now().Add(-24 * time.Hour)
	return inv
}

func NewTestOutboxEvent() *entity.OutboxEvent {
	now := time.Now()
	return &entity.OutboxEvent{
		ID:          uuid.New(),
		EventType:   "test.event",
		Payload:     `{"key": "value"}`,
		Status:      "pending",
		RetryCount:  0,
		NextRetryAt: nil,
		LastErrorAt: nil,
		ProcessedAt: nil,
		CreatedAt:   now,
	}
}

func NewTestOutboxEventWithType(eventType string) *entity.OutboxEvent {
	ev := NewTestOutboxEvent()
	ev.EventType = eventType
	return ev
}

func NewTestOutboxEventWithPayload(payload string) *entity.OutboxEvent {
	ev := NewTestOutboxEvent()
	ev.Payload = payload
	return ev
}
