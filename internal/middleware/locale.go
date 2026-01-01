package middleware

import (
	sharedmw "github.com/creafly/middleware"
	"github.com/creafly/notifications/internal/i18n"
	"github.com/gin-gonic/gin"
)

func GetLocale(c *gin.Context) i18n.Locale {
	return i18n.ParseLocale(sharedmw.GetLocale(c))
}
