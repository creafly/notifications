package handler

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/creafly/notifications/internal/domain/entity"
	"github.com/creafly/notifications/internal/domain/service"
	"github.com/creafly/notifications/internal/i18n"
	"github.com/creafly/notifications/internal/infra/client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PushNotificationHandler struct {
	pushService    service.PushNotificationService
	identityClient *client.IdentityClient
}

func NewPushNotificationHandler(pushService service.PushNotificationService, identityClient *client.IdentityClient) *PushNotificationHandler {
	return &PushNotificationHandler{
		pushService:    pushService,
		identityClient: identityClient,
	}
}

type CreatePushRequest struct {
	Title          string              `json:"title" binding:"required"`
	Message        string              `json:"message" binding:"required"`
	TargetType     string              `json:"targetType" binding:"required,oneof=all tenant users"`
	TargetTenantID *string             `json:"targetTenantId,omitempty"`
	TargetUserIDs  []string            `json:"targetUserIds,omitempty"`
	Buttons        []entity.PushButton `json:"buttons,omitempty"`
	ScheduledAt    *string             `json:"scheduledAt,omitempty"`
	SendNow        bool                `json:"sendNow,omitempty"`
}

type UpdatePushRequest struct {
	Title          *string             `json:"title,omitempty"`
	Message        *string             `json:"message,omitempty"`
	TargetType     *string             `json:"targetType,omitempty"`
	TargetTenantID *string             `json:"targetTenantId,omitempty"`
	TargetUserIDs  []string            `json:"targetUserIds,omitempty"`
	Buttons        []entity.PushButton `json:"buttons,omitempty"`
	ScheduledAt    *string             `json:"scheduledAt,omitempty"`
}

type SendPushRequest struct {
	UserIDs []string `json:"userIds"`
}

func (h *PushNotificationHandler) Create(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}
	createdBy := userIDVal.(uuid.UUID)

	var req CreatePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	input := service.CreatePushInput{
		Title:      req.Title,
		Message:    req.Message,
		TargetType: entity.PushTargetType(req.TargetType),
		Buttons:    req.Buttons,
		CreatedBy:  createdBy,
	}

	if req.TargetTenantID != nil {
		tenantID, err := uuid.Parse(*req.TargetTenantID)
		if err == nil {
			input.TargetTenantID = &tenantID
		}
	}

	for _, idStr := range req.TargetUserIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			input.TargetUserIDs = append(input.TargetUserIDs, id)
		}
	}

	if req.ScheduledAt != nil {
		scheduledAt, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err == nil {
			input.ScheduledAt = &scheduledAt
		}
	}

	push, err := h.pushService.Create(c.Request.Context(), input)
	if err != nil {
		log.Printf("[ERROR] Failed to create push notification: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	if req.SendNow && len(input.TargetUserIDs) > 0 {
		if err := h.pushService.Send(c.Request.Context(), push.ID, input.TargetUserIDs); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
			return
		}
		push, _ = h.pushService.GetByID(c.Request.Context(), push.ID)
	}

	c.JSON(http.StatusCreated, push)
}

func (h *PushNotificationHandler) GetAll(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	pushes, total, err := h.pushService.GetAll(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   pushes,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *PushNotificationHandler) GetByID(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	push, err := h.pushService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
		return
	}

	c.JSON(http.StatusOK, push)
}

func (h *PushNotificationHandler) Update(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	var req UpdatePushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	input := service.UpdatePushInput{
		Title:   req.Title,
		Message: req.Message,
		Buttons: req.Buttons,
	}

	if req.TargetType != nil {
		targetType := entity.PushTargetType(*req.TargetType)
		input.TargetType = &targetType
	}

	if req.TargetTenantID != nil {
		tenantID, err := uuid.Parse(*req.TargetTenantID)
		if err == nil {
			input.TargetTenantID = &tenantID
		}
	}

	for _, idStr := range req.TargetUserIDs {
		if id, err := uuid.Parse(idStr); err == nil {
			input.TargetUserIDs = append(input.TargetUserIDs, id)
		}
	}

	if req.ScheduledAt != nil {
		scheduledAt, err := time.Parse(time.RFC3339, *req.ScheduledAt)
		if err == nil {
			input.ScheduledAt = &scheduledAt
		}
	}

	push, err := h.pushService.Update(c.Request.Context(), id, input)
	if err != nil {
		if err == service.ErrPushAlreadySent {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Push.AlreadySent})
			return
		}
		if err == service.ErrPushNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, push)
}

func (h *PushNotificationHandler) Delete(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.pushService.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrPushAlreadySent {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Push.AlreadySent})
			return
		}
		if err == service.ErrPushNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Push.Deleted})
}

func (h *PushNotificationHandler) Send(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	push, err := h.pushService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
		return
	}

	accessToken := ""
	if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 {
		accessToken = authHeader[7:]
	}

	var userIDs []uuid.UUID

	var req SendPushRequest
	if err := c.ShouldBindJSON(&req); err == nil && len(req.UserIDs) > 0 {
		for _, idStr := range req.UserIDs {
			if uid, err := uuid.Parse(idStr); err == nil {
				userIDs = append(userIDs, uid)
			}
		}
	}

	if len(userIDs) == 0 {
		switch push.TargetType {
		case entity.PushTargetAll:
			if h.identityClient != nil && accessToken != "" {
				fetchedIDs, err := h.identityClient.GetAllUsers(c.Request.Context(), accessToken)
				if err == nil {
					userIDs = fetchedIDs
				}
			}
		case entity.PushTargetTenant:
			if h.identityClient != nil && accessToken != "" && push.TargetTenantID != nil {
				fetchedIDs, err := h.identityClient.GetTenantMembers(c.Request.Context(), accessToken, *push.TargetTenantID)
				if err == nil {
					userIDs = fetchedIDs
				}
			}
		case entity.PushTargetUsers:
			userIDs = push.TargetUserIDs
		}
	}

	if len(userIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Push.NoRecipients})
		return
	}

	if err := h.pushService.Send(c.Request.Context(), id, userIDs); err != nil {
		if err == service.ErrPushAlreadySent {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Push.AlreadySent})
			return
		}
		if err == service.ErrPushNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Push.Sent})
}

func (h *PushNotificationHandler) Cancel(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.pushService.Cancel(c.Request.Context(), id); err != nil {
		if err == service.ErrPushAlreadySent {
			c.JSON(http.StatusBadRequest, gin.H{"error": messages.Push.AlreadySent})
			return
		}
		if err == service.ErrPushNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Push.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Push.Cancelled})
}

func (h *PushNotificationHandler) GetMyPushNotifications(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}
	userID := userIDVal.(uuid.UUID)

	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	pushes, err := h.pushService.GetUserPushNotifications(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, pushes)
}

func (h *PushNotificationHandler) MarkAsRead(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}
	userID := userIDVal.(uuid.UUID)

	pushID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.pushService.MarkAsRead(c.Request.Context(), pushID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Push.MarkedRead})
}
