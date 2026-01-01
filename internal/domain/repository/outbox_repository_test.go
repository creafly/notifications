package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hexaend/notifications/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutboxRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, event.ID, events[0].ID)
		assert.Equal(t, event.EventType, events[0].EventType)
		assert.Equal(t, event.Payload, events[0].Payload)
	})

	t.Run("with custom event type", func(t *testing.T) {
		event := testutil.NewTestOutboxEventWithType("notification.created")
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		found := false
		for _, e := range events {
			if e.ID == event.ID {
				assert.Equal(t, "notification.created", e.EventType)
				found = true
				break
			}
		}
		assert.True(t, found)
	})
}

func TestOutboxRepository_GetPending(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)
	}

	t.Run("returns pending events", func(t *testing.T) {
		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		assert.Len(t, events, 5)
	})

	t.Run("respects limit", func(t *testing.T) {
		events, err := repo.GetPending(ctx, 2)
		require.NoError(t, err)
		assert.Len(t, events, 2)
	})
}

func TestOutboxRepository_MarkAsProcessed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		err = repo.MarkAsProcessed(ctx, event.ID)
		require.NoError(t, err)

		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		for _, e := range events {
			assert.NotEqual(t, event.ID, e.ID)
		}
	})
}

func TestOutboxRepository_MarkAsFailed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		err = repo.MarkAsFailed(ctx, event.ID)
		require.NoError(t, err)

		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		for _, e := range events {
			assert.NotEqual(t, event.ID, e.ID)
		}
	})
}

func TestOutboxRepository_IncrementRetry(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		nextRetry := time.Now().Add(5 * time.Minute)
		err = repo.IncrementRetry(ctx, event.ID, nextRetry)
		require.NoError(t, err)
	})
}

func TestOutboxRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		event := testutil.NewTestOutboxEvent()
		err := repo.Create(ctx, event)
		require.NoError(t, err)

		err = repo.Delete(ctx, event.ID)
		require.NoError(t, err)

		events, err := repo.GetPending(ctx, 10)
		require.NoError(t, err)
		for _, e := range events {
			assert.NotEqual(t, event.ID, e.ID)
		}
	})

	t.Run("non-existent event", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New())
		assert.NoError(t, err)
	})
}

func TestOutboxRepository_CleanupOld(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	event := testutil.NewTestOutboxEvent()
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	err = repo.MarkAsProcessed(ctx, event.ID)
	require.NoError(t, err)

	t.Run("does not cleanup recent", func(t *testing.T) {
		err := repo.CleanupOld(ctx, 24*time.Hour)
		require.NoError(t, err)
	})
}

func TestOutboxRepository_CleanupFailed(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewOutboxRepository(tdb.DB)
	ctx := context.Background()

	event := testutil.NewTestOutboxEvent()
	err := repo.Create(ctx, event)
	require.NoError(t, err)

	err = repo.MarkAsFailed(ctx, event.ID)
	require.NoError(t, err)

	t.Run("does not cleanup recent", func(t *testing.T) {
		err := repo.CleanupFailed(ctx, 24*time.Hour)
		require.NoError(t, err)
	})
}
