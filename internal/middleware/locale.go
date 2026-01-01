package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hexaend/notifications/internal/i18n"
)

func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := i18n.LocaleEnUS

		acceptLang := c.GetHeader("Accept-Language")
		if acceptLang != "" {
			if strings.Contains(strings.ToLower(acceptLang), "ru") {
				locale = i18n.LocaleRuRU
			}
		}

		c.Set("locale", locale)
		c.Next()
	}
}

func GetLocale(c *gin.Context) i18n.Locale {
	if locale, exists := c.Get("locale"); exists {
		if l, ok := locale.(i18n.Locale); ok {
			return l
		}
	}
	return i18n.LocaleEnUS
}
