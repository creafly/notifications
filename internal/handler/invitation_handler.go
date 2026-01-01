package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/creafly/notifications/internal/domain/service"
	"github.com/creafly/notifications/internal/i18n"
)

type InvitationHandler struct {
	invitationService service.InvitationService
}

func NewInvitationHandler(invitationService service.InvitationService) *InvitationHandler {
	return &InvitationHandler{invitationService: invitationService}
}

func getLocale(c *gin.Context) i18n.Locale {
	if locale, exists := c.Get("locale"); exists {
		if l, ok := locale.(i18n.Locale); ok {
			return l
		}
	}
	return i18n.LocaleEnUS
}

func (h *InvitationHandler) GetMyInvitations(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	userID := userIDVal.(uuid.UUID)

	invitations, err := h.invitationService.GetPendingByInviteeID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

func (h *InvitationHandler) GetByTenant(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	tenantID, err := uuid.Parse(c.Param("tenantId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	invitations, err := h.invitationService.GetByTenantID(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

func (h *InvitationHandler) Accept(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.invitationService.Accept(c.Request.Context(), id); err != nil {
		if err == service.ErrInvitationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Invitation.NotFound})
			return
		}
		if err == service.ErrInvitationExpired {
			c.JSON(http.StatusGone, gin.H{"error": messages.Invitation.Expired})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Invitation.Accepted})
}

func (h *InvitationHandler) Reject(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	if err := h.invitationService.Reject(c.Request.Context(), id); err != nil {
		if err == service.ErrInvitationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": messages.Invitation.NotFound})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": messages.Invitation.Rejected})
}

type CreateInvitationRequest struct {
	TenantID   string  `json:"tenantId" binding:"required,uuid"`
	TenantName string  `json:"tenantName" binding:"required"`
	InviteeID  string  `json:"inviteeId" binding:"required,uuid"`
	Email      string  `json:"email" binding:"required,email"`
	RoleID     *string `json:"roleId"`
}

func (h *InvitationHandler) Create(c *gin.Context) {
	locale := getLocale(c)
	messages := i18n.GetMessages(locale)

	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
		return
	}

	userID := userIDVal.(uuid.UUID)

	inviterName := "User"
	if name, exists := c.Get("userName"); exists {
		if n, ok := name.(string); ok {
			inviterName = n
		}
	}

	var req CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": messages.Errors.ValidationFailed})
		return
	}

	tenantID, _ := uuid.Parse(req.TenantID)
	inviteeID, _ := uuid.Parse(req.InviteeID)

	var roleID *uuid.UUID
	if req.RoleID != nil {
		id, err := uuid.Parse(*req.RoleID)
		if err == nil {
			roleID = &id
		}
	}

	invitation, err := h.invitationService.Create(c.Request.Context(), service.CreateInvitationInput{
		TenantID:    tenantID,
		TenantName:  req.TenantName,
		InviterID:   userID,
		InviterName: inviterName,
		InviteeID:   inviteeID,
		Email:       req.Email,
		RoleID:      roleID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": messages.Errors.InternalError})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    messages.Invitation.Created,
		"invitation": invitation,
	})
}
