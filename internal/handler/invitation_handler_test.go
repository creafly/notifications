package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/domain/service"
	"github.com/creafly/notifications/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type invitationServiceMock struct {
	CreateFunc                func(ctx context.Context, input service.CreateInvitationInput) (*entity.Invitation, error)
	GetByIDFunc               func(ctx context.Context, id uuid.UUID) (*entity.Invitation, error)
	GetByInviteeIDFunc        func(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetPendingByInviteeIDFunc func(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error)
	GetByTenantIDFunc         func(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error)
	AcceptFunc                func(ctx context.Context, id uuid.UUID) error
	RejectFunc                func(ctx context.Context, id uuid.UUID) error
	CancelFunc                func(ctx context.Context, id uuid.UUID) error
}

func (m *invitationServiceMock) Create(ctx context.Context, input service.CreateInvitationInput) (*entity.Invitation, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}

func (m *invitationServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Invitation, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *invitationServiceMock) GetByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	if m.GetByInviteeIDFunc != nil {
		return m.GetByInviteeIDFunc(ctx, inviteeID)
	}
	return nil, nil
}

func (m *invitationServiceMock) GetPendingByInviteeID(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
	if m.GetPendingByInviteeIDFunc != nil {
		return m.GetPendingByInviteeIDFunc(ctx, inviteeID)
	}
	return nil, nil
}

func (m *invitationServiceMock) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*entity.Invitation, error) {
	if m.GetByTenantIDFunc != nil {
		return m.GetByTenantIDFunc(ctx, tenantID)
	}
	return nil, nil
}

func (m *invitationServiceMock) Accept(ctx context.Context, id uuid.UUID) error {
	if m.AcceptFunc != nil {
		return m.AcceptFunc(ctx, id)
	}
	return nil
}

func (m *invitationServiceMock) Reject(ctx context.Context, id uuid.UUID) error {
	if m.RejectFunc != nil {
		return m.RejectFunc(ctx, id)
	}
	return nil
}

func (m *invitationServiceMock) Cancel(ctx context.Context, id uuid.UUID) error {
	if m.CancelFunc != nil {
		return m.CancelFunc(ctx, id)
	}
	return nil
}

func setupInvitationRouter(svc service.InvitationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewInvitationHandler(svc)

	router.GET("/invitations", handler.GetMyInvitations)
	router.GET("/invitations/tenant/:tenantId", handler.GetByTenant)
	router.POST("/invitations", handler.Create)
	router.POST("/invitations/:id/accept", handler.Accept)
	router.POST("/invitations/:id/reject", handler.Reject)

	return router
}

func TestInvitationHandler_GetMyInvitations(t *testing.T) {
	userID := utils.GenerateUUID()

	t.Run("unauthorized", func(t *testing.T) {
		svc := &invitationServiceMock{}
		router := setupInvitationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/invitations", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("with context user", func(t *testing.T) {
		svc := &invitationServiceMock{
			GetPendingByInviteeIDFunc: func(ctx context.Context, inviteeID uuid.UUID) ([]*entity.Invitation, error) {
				return []*entity.Invitation{
					{ID: utils.GenerateUUID(), InviteeID: inviteeID},
				}, nil
			},
		}

		gin.SetMode(gin.TestMode)
		router := gin.New()
		handler := NewInvitationHandler(svc)

		router.GET("/invitations", func(c *gin.Context) {
			c.Set("userID", userID)
			handler.GetMyInvitations(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/invitations", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestInvitationHandler_GetByTenant(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tenantID := utils.GenerateUUID()
		svc := &invitationServiceMock{
			GetByTenantIDFunc: func(ctx context.Context, tid uuid.UUID) ([]*entity.Invitation, error) {
				return []*entity.Invitation{
					{ID: utils.GenerateUUID(), TenantID: tid},
				}, nil
			},
		}
		router := setupInvitationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/invitations/tenant/"+tenantID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		svc := &invitationServiceMock{}
		router := setupInvitationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/invitations/tenant/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestInvitationHandler_Accept(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &invitationServiceMock{
			AcceptFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		router := setupInvitationRouter(svc)

		invitationID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodPost, "/invitations/"+invitationID.String()+"/accept", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &invitationServiceMock{
			AcceptFunc: func(ctx context.Context, id uuid.UUID) error {
				return service.ErrInvitationNotFound
			},
		}
		router := setupInvitationRouter(svc)

		invitationID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodPost, "/invitations/"+invitationID.String()+"/accept", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("expired", func(t *testing.T) {
		svc := &invitationServiceMock{
			AcceptFunc: func(ctx context.Context, id uuid.UUID) error {
				return service.ErrInvitationExpired
			},
		}
		router := setupInvitationRouter(svc)

		invitationID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodPost, "/invitations/"+invitationID.String()+"/accept", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusGone, w.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		svc := &invitationServiceMock{}
		router := setupInvitationRouter(svc)

		req := httptest.NewRequest(http.MethodPost, "/invitations/invalid-uuid/accept", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestInvitationHandler_Reject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &invitationServiceMock{
			RejectFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		router := setupInvitationRouter(svc)

		invitationID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodPost, "/invitations/"+invitationID.String()+"/reject", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		svc := &invitationServiceMock{
			RejectFunc: func(ctx context.Context, id uuid.UUID) error {
				return service.ErrInvitationNotFound
			},
		}
		router := setupInvitationRouter(svc)

		invitationID := utils.GenerateUUID()
		req := httptest.NewRequest(http.MethodPost, "/invitations/"+invitationID.String()+"/reject", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestInvitationHandler_Create(t *testing.T) {
	userID := utils.GenerateUUID()

	t.Run("unauthorized", func(t *testing.T) {
		svc := &invitationServiceMock{}
		router := setupInvitationRouter(svc)

		body := `{"tenantId":"` + utils.GenerateUUID().String() + `","tenantName":"Test","inviteeId":"` + utils.GenerateUUID().String() + `","email":"test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/invitations", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("success", func(t *testing.T) {
		svc := &invitationServiceMock{
			CreateFunc: func(ctx context.Context, input service.CreateInvitationInput) (*entity.Invitation, error) {
				return &entity.Invitation{
					ID:       utils.GenerateUUID(),
					TenantID: input.TenantID,
					Email:    input.Email,
				}, nil
			},
		}

		gin.SetMode(gin.TestMode)
		router := gin.New()
		handler := NewInvitationHandler(svc)

		router.POST("/invitations", func(c *gin.Context) {
			c.Set("userID", userID)
			c.Set("userName", "Test User")
			handler.Create(c)
		})

		body := `{"tenantId":"` + utils.GenerateUUID().String() + `","tenantName":"Test","inviteeId":"` + utils.GenerateUUID().String() + `","email":"test@example.com"}`
		req := httptest.NewRequest(http.MethodPost, "/invitations", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response["invitation"])
	})

	t.Run("validation error", func(t *testing.T) {
		svc := &invitationServiceMock{}

		gin.SetMode(gin.TestMode)
		router := gin.New()
		handler := NewInvitationHandler(svc)

		router.POST("/invitations", func(c *gin.Context) {
			c.Set("userID", userID)
			handler.Create(c)
		})

		body := `{"tenantName":"Test"}`
		req := httptest.NewRequest(http.MethodPost, "/invitations", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
