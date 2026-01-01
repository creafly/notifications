package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hexaend/notifications/internal/domain/entity"
	"github.com/hexaend/notifications/internal/domain/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notificationServiceMock struct {
	CreateFunc            func(ctx context.Context, input service.CreateNotificationInput) (*entity.Notification, error)
	GetByIDFunc           func(ctx context.Context, id uuid.UUID) (*entity.Notification, error)
	GetByUserIDFunc       func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error)
	GetUnreadByUserIDFunc func(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error)
	GetUnreadCountFunc    func(ctx context.Context, userID uuid.UUID) (int, error)
	MarkAsReadFunc        func(ctx context.Context, id uuid.UUID) error
	MarkAllAsReadFunc     func(ctx context.Context, userID uuid.UUID) error
	DeleteFunc            func(ctx context.Context, id uuid.UUID) error
}

func (m *notificationServiceMock) Create(ctx context.Context, input service.CreateNotificationInput) (*entity.Notification, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, input)
	}
	return nil, nil
}

func (m *notificationServiceMock) GetByID(ctx context.Context, id uuid.UUID) (*entity.Notification, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *notificationServiceMock) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *notificationServiceMock) GetUnreadByUserID(ctx context.Context, userID uuid.UUID) ([]*entity.Notification, error) {
	if m.GetUnreadByUserIDFunc != nil {
		return m.GetUnreadByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *notificationServiceMock) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	if m.GetUnreadCountFunc != nil {
		return m.GetUnreadCountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *notificationServiceMock) MarkAsRead(ctx context.Context, id uuid.UUID) error {
	if m.MarkAsReadFunc != nil {
		return m.MarkAsReadFunc(ctx, id)
	}
	return nil
}

func (m *notificationServiceMock) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	if m.MarkAllAsReadFunc != nil {
		return m.MarkAllAsReadFunc(ctx, userID)
	}
	return nil
}

func (m *notificationServiceMock) Delete(ctx context.Context, id uuid.UUID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func setupNotificationRouter(svc service.NotificationService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewNotificationHandler(svc)

	router.GET("/notifications", handler.GetMyNotifications)
	router.GET("/notifications/unread", handler.GetUnreadNotifications)
	router.GET("/notifications/unread/count", handler.GetUnreadCount)
	router.POST("/notifications/:id/read", handler.MarkAsRead)
	router.POST("/notifications/read-all", handler.MarkAllAsRead)
	router.DELETE("/notifications/:id", handler.Delete)

	return router
}

func TestNotificationHandler_GetMyNotifications(t *testing.T) {
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		svc := &notificationServiceMock{
			GetByUserIDFunc: func(ctx context.Context, uid uuid.UUID, limit, offset int) ([]*entity.Notification, error) {
				return []*entity.Notification{
					{ID: uuid.New(), UserID: uid},
				}, nil
			},
		}
		router := setupNotificationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		w := httptest.NewRecorder()

		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Set("userID", userID)

		router.ServeHTTP(w, req)
	})

	t.Run("unauthorized", func(t *testing.T) {
		svc := &notificationServiceMock{}
		router := setupNotificationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestNotificationHandler_GetUnreadCount(t *testing.T) {
	userID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		svc := &notificationServiceMock{}
		router := setupNotificationRouter(svc)

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("with context user", func(t *testing.T) {
		svc := &notificationServiceMock{
			GetUnreadCountFunc: func(ctx context.Context, uid uuid.UUID) (int, error) {
				return 5, nil
			},
		}

		gin.SetMode(gin.TestMode)
		router := gin.New()
		handler := NewNotificationHandler(svc)

		router.GET("/notifications/unread/count", func(c *gin.Context) {
			c.Set("userID", userID)
			handler.GetUnreadCount(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/notifications/unread/count", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]int
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, 5, response["count"])
	})
}

func TestNotificationHandler_MarkAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &notificationServiceMock{
			MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		router := setupNotificationRouter(svc)

		notificationID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/notifications/"+notificationID.String()+"/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		svc := &notificationServiceMock{}
		router := setupNotificationRouter(svc)

		req := httptest.NewRequest(http.MethodPost, "/notifications/invalid-uuid/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("service error", func(t *testing.T) {
		svc := &notificationServiceMock{
			MarkAsReadFunc: func(ctx context.Context, id uuid.UUID) error {
				return errors.New("db error")
			},
		}
		router := setupNotificationRouter(svc)

		notificationID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/notifications/"+notificationID.String()+"/read", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNotificationHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := &notificationServiceMock{
			DeleteFunc: func(ctx context.Context, id uuid.UUID) error {
				return nil
			},
		}
		router := setupNotificationRouter(svc)

		notificationID := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/notifications/"+notificationID.String(), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		svc := &notificationServiceMock{}
		router := setupNotificationRouter(svc)

		req := httptest.NewRequest(http.MethodDelete, "/notifications/invalid-uuid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
