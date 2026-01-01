package middleware

import (
	"net/http"

	"github.com/creafly/notifications/internal/i18n"
	"github.com/creafly/notifications/internal/infra/client"
	"github.com/gin-gonic/gin"
)

func RequireAnyClaim(identityClient *client.IdentityClient, claims ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		userClaims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			return
		}

		claimsList, ok := userClaims.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			return
		}

		hasAnyClaim := false
		for _, requiredClaim := range claims {
			for _, userClaim := range claimsList {
				if userClaim == requiredClaim {
					hasAnyClaim = true
					break
				}
			}
			if hasAnyClaim {
				break
			}
		}

		if !hasAnyClaim {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			return
		}

		c.Next()
	}
}

func RequireAllClaims(identityClient *client.IdentityClient, claims ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := GetLocale(c)
		messages := i18n.GetMessages(locale)

		userClaims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			return
		}

		claimsList, ok := userClaims.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
			return
		}

		claimsMap := make(map[string]bool)
		for _, claim := range claimsList {
			claimsMap[claim] = true
		}

		for _, requiredClaim := range claims {
			if !claimsMap[requiredClaim] {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": messages.Errors.Forbidden})
				return
			}
		}

		c.Next()
	}
}

func HasClaim(c *gin.Context, claim string) bool {
	userClaims, exists := c.Get("claims")
	if !exists {
		return false
	}

	claimsList, ok := userClaims.([]string)
	if !ok {
		return false
	}

	for _, userClaim := range claimsList {
		if userClaim == claim {
			return true
		}
	}

	return false
}

func GetUserClaims(c *gin.Context) []string {
	userClaims, exists := c.Get("claims")
	if !exists {
		return nil
	}

	claimsList, ok := userClaims.([]string)
	if !ok {
		return nil
	}

	return claimsList
}
