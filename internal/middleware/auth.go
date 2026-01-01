package middleware

import (
	"net/http"
	"strings"

	"github.com/creafly/notifications/internal/i18n"
	"github.com/creafly/notifications/internal/infra/client"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(identityClient *client.IdentityClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		var tokenString string

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				tokenString = ""
			}
		}

		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			return
		}

		verifyResp, err := identityClient.VerifyToken(c.Request.Context(), tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			return
		}

		if !verifyResp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": messages.Errors.Unauthorized})
			return
		}

		if verifyResp.IsBlocked {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":       messages.Errors.UserBlocked,
				"blockReason": verifyResp.BlockReason,
				"blockedAt":   verifyResp.BlockedAt,
			})
			return
		}

		c.Set("userID", verifyResp.UserID)
		c.Set("email", verifyResp.Email)
		c.Set("claims", verifyResp.Claims)
		if verifyResp.TenantID != nil {
			c.Set("tenantID", *verifyResp.TenantID)
		}

		c.Next()
	}
}
