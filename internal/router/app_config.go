package router

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/config"
)

func appConfigHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, _ := json.Marshal(gin.H{"adminRoutePrefix": cfg.AdminRoutePrefix})
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.String(http.StatusOK, "window.__APP_CONFIG__ = %s;", payload)
	}
}
