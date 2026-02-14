package repository

import (
	"context"
	"testing"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/testutil"
	"github.com/creafly/notifications/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvitationRepository_Create(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, invitation.ID)
		require.NoError(t, err)
		assert.Equal(t, invitation.ID, result.ID)
		assert.Equal(t, invitation.TenantID, result.TenantID)
		assert.Equal(t, invitation.TenantName, result.TenantName)
		assert.Equal(t, invitation.InviterID, result.InviterID)
		assert.Equal(t, invitation.InviteeID, result.InviteeID)
		assert.Equal(t, invitation.Email, result.Email)
		assert.Equal(t, invitation.Status, result.Status)
	})

	t.Run("with role", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		roleID := utils.GenerateUUID()
		invitation.RoleID = &roleID
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, invitation.ID)
		require.NoError(t, err)
		require.NotNil(t, result.RoleID)
		assert.Equal(t, roleID, *result.RoleID)
	})
}

func TestInvitationRepository_GetByID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, invitation.ID)
		require.NoError(t, err)
		assert.Equal(t, invitation.ID, result.ID)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, utils.GenerateUUID())
		assert.Error(t, err)
	})
}

func TestInvitationRepository_GetByInviteeID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	inviteeID := utils.GenerateUUID()

	for i := 0; i < 4; i++ {
		invitation := testutil.NewTestInvitationWithInvitee(inviteeID)
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		invitation := testutil.NewTestInvitation()
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)
	}

	t.Run("returns invitee invitations", func(t *testing.T) {
		results, err := repo.GetByInviteeID(ctx, inviteeID)
		require.NoError(t, err)
		assert.Len(t, results, 4)
		for _, inv := range results {
			assert.Equal(t, inviteeID, inv.InviteeID)
		}
	})

	t.Run("empty for unknown invitee", func(t *testing.T) {
		results, err := repo.GetByInviteeID(ctx, utils.GenerateUUID())
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestInvitationRepository_GetPendingByInviteeID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	inviteeID := utils.GenerateUUID()

	for i := 0; i < 2; i++ {
		invitation := testutil.NewTestInvitationWithInvitee(inviteeID)
		invitation.Status = entity.InvitationStatusPending
		invitation.ExpiresAt = time.Now().Add(24 * time.Hour)
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)
	}

	accepted := testutil.NewTestInvitationWithInvitee(inviteeID)
	accepted.Status = entity.InvitationStatusAccepted
	err := repo.Create(ctx, accepted)
	require.NoError(t, err)

	expired := testutil.NewTestInvitationWithInvitee(inviteeID)
	expired.Status = entity.InvitationStatusPending
	expired.ExpiresAt = time.Now().Add(-24 * time.Hour)
	err = repo.Create(ctx, expired)
	require.NoError(t, err)

	t.Run("returns only pending non-expired", func(t *testing.T) {
		results, err := repo.GetPendingByInviteeID(ctx, inviteeID)
		require.NoError(t, err)
		assert.Len(t, results, 2)
		for _, inv := range results {
			assert.Equal(t, entity.InvitationStatusPending, inv.Status)
			assert.True(t, inv.ExpiresAt.After(time.Now()))
		}
	})
}

func TestInvitationRepository_GetByTenantID(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	tenantID := utils.GenerateUUID()

	for i := 0; i < 3; i++ {
		invitation := testutil.NewTestInvitationWithTenant(tenantID)
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)
	}

	for i := 0; i < 2; i++ {
		invitation := testutil.NewTestInvitation()
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)
	}

	t.Run("returns tenant invitations", func(t *testing.T) {
		results, err := repo.GetByTenantID(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, results, 3)
		for _, inv := range results {
			assert.Equal(t, tenantID, inv.TenantID)
		}
	})

	t.Run("empty for unknown tenant", func(t *testing.T) {
		results, err := repo.GetByTenantID(ctx, utils.GenerateUUID())
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}

func TestInvitationRepository_UpdateStatus(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		invitation.Status = entity.InvitationStatusPending
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, invitation.ID, entity.InvitationStatusAccepted)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, invitation.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.InvitationStatusAccepted, result.Status)
	})

	t.Run("reject invitation", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		invitation.Status = entity.InvitationStatusPending
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		err = repo.UpdateStatus(ctx, invitation.ID, entity.InvitationStatusRejected)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, invitation.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.InvitationStatusRejected, result.Status)
	})
}

func TestInvitationRepository_Delete(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		invitation := testutil.NewTestInvitation()
		err := repo.Create(ctx, invitation)
		require.NoError(t, err)

		err = repo.Delete(ctx, invitation.ID)
		require.NoError(t, err)

		_, err = repo.GetByID(ctx, invitation.ID)
		assert.Error(t, err)
	})

	t.Run("non-existent invitation", func(t *testing.T) {
		err := repo.Delete(ctx, utils.GenerateUUID())
		assert.NoError(t, err)
	})
}

func TestInvitationRepository_ExpireOld(t *testing.T) {
	tdb := testutil.SetupTestDB(t)
	defer tdb.Cleanup(t)

	repo := NewInvitationRepository(tdb.DB)
	ctx := context.Background()

	expired1 := testutil.NewTestInvitation()
	expired1.Status = entity.InvitationStatusPending
	expired1.ExpiresAt = time.Now().Add(-48 * time.Hour)
	err := repo.Create(ctx, expired1)
	require.NoError(t, err)

	expired2 := testutil.NewTestInvitation()
	expired2.Status = entity.InvitationStatusPending
	expired2.ExpiresAt = time.Now().Add(-24 * time.Hour)
	err = repo.Create(ctx, expired2)
	require.NoError(t, err)

	valid := testutil.NewTestInvitation()
	valid.Status = entity.InvitationStatusPending
	valid.ExpiresAt = time.Now().Add(24 * time.Hour)
	err = repo.Create(ctx, valid)
	require.NoError(t, err)

	t.Run("expires old invitations", func(t *testing.T) {
		err := repo.ExpireOld(ctx)
		require.NoError(t, err)

		result1, err := repo.GetByID(ctx, expired1.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.InvitationStatusExpired, result1.Status)

		result2, err := repo.GetByID(ctx, expired2.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.InvitationStatusExpired, result2.Status)

		resultValid, err := repo.GetByID(ctx, valid.ID)
		require.NoError(t, err)
		assert.Equal(t, entity.InvitationStatusPending, resultValid.Status)
	})
}
