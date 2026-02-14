package repository

import (
	"context"
	"testing"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/testutil"
	"github.com/creafly/notifications/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		notification := testutil.NewTestNotification()
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.ID, result.ID)
		assert.Equal(t, notification.UserID, result.UserID)
		assert.Equal(t, notification.Title, result.Title)
		assert.Equal(t, notification.Type, result.Type)
		assert.Equal(t, notification.Status, result.Status)
	})

	t.Run("with tenant", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		notification := testutil.NewTestNotificationWithTenant(utils.GenerateUUID(), tenantID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.NotNil(t, result.TenantID)
		assert.Equal(t, tenantID, *result.TenantID)
	})

	t.Run("with data", func(t *testing.T) {
		notification := testutil.NewTestNotification()
		data := `{"key": "value"}`
		notification.Data = &data
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		require.NotNil(t, result.Data)
		assert.Equal(t, data, *result.Data)
	})
}

func TestNotificationRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		notification := testutil.NewTestNotification()
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, notification.ID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, utils.GenerateUUID())
		assert.Error(t, err)
	})
}

func TestNotificationRepository_GetByUserID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	userID := utils.GenerateUUID()

	for i := 0; i < 5; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	for i := 0; i < 3; i++ {
		notification := testutil.NewTestNotification()
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	t.Run("returns user notifications", func(t *testing.T) {
		results, err := repo.GetByUserID(ctx, userID, 10, 0)
		require.NoError(t, err)
		assert.Len(t, results, 5)
		for _, n := range results {
			assert.Equal(t, userID, n.UserID)
		}
	})

	t.Run("with limit", func(t *testing.T) {
		results, err := repo.GetByUserID(ctx, userID, 2, 0)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("with offset", func(t *testing.T) {
		results, err := repo.GetByUserID(ctx, userID, 10, 3)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("empty for unknown user", func(t *testing.T) {
		results, err := repo.GetByUserID(ctx, utils.GenerateUUID(), 10, 0)
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestNotificationRepository_GetUnreadByUserID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	userID := utils.GenerateUUID()

	for i := 0; i < 3; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		notification.Status = entity.NotificationStatusUnread
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		notification.Status = entity.NotificationStatusRead
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	t.Run("returns only unread", func(t *testing.T) {
		results, err := repo.GetUnreadByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		for _, n := range results {
			assert.Equal(t, entity.NotificationStatusUnread, n.Status)
		}
	})
}

func TestNotificationRepository_GetUnreadCount(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	userID := utils.GenerateUUID()

	for i := 0; i < 4; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		notification.Status = entity.NotificationStatusUnread
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		notification.Status = entity.NotificationStatusRead
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	t.Run("returns correct count", func(t *testing.T) {
		count, err := repo.GetUnreadCount(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, 4, count)
	})

	t.Run("zero for user without notifications", func(t *testing.T) {
		count, err := repo.GetUnreadCount(ctx, utils.GenerateUUID())
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestNotificationRepository_MarkAsRead(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		notification := testutil.NewTestNotification()
		notification.Status = entity.NotificationStatusUnread
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		err = repo.MarkAsRead(ctx, notification.ID)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, notification.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.NotificationStatusRead, result.Status)
		assert.NotNil(t, result.ReadAt)
	})

	t.Run("non-existent notification", func(t *testing.T) {
		err := repo.MarkAsRead(ctx, utils.GenerateUUID())
		assert.NoError(t, err)
	})
}

func TestNotificationRepository_MarkAllAsRead(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	userID := utils.GenerateUUID()

	for i := 0; i < 3; i++ {
		notification := testutil.NewTestNotificationWithUser(userID)
		notification.Status = entity.NotificationStatusUnread
		err := repo.Create(ctx, notification)
		require.NoError(t, err)
	}

	t.Run("marks all as read", func(t *testing.T) {
		err := repo.MarkAllAsRead(ctx, userID)
		require.NoError(t, err)

		unread, err := repo.GetUnreadByUserID(ctx, userID)
		require.NoError(t, err)
		assert.Empty(t, unread)
	})
}

func TestNotificationRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewNotificationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		notification := testutil.NewTestNotification()
		err := repo.Create(ctx, notification)
		require.NoError(t, err)

		err = repo.Delete(ctx, notification.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, notification.ID)
		assert.Error(t, err)
	})

	t.Run("non-existent notification", func(t *testing.T) {
		err := repo.Delete(ctx, utils.GenerateUUID())
		assert.NoError(t, err)
	})
}
