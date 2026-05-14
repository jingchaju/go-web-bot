package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/bot"
	"go-web-bot/internal/config"
)

func registerFrontend(r *gin.Engine, cfg config.Config) {
	dist := cfg.FrontendDist
	indexPath := filepath.Join(dist, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "telegram bot admin service") })
		return
	}

	r.GET("/", func(c *gin.Context) { c.File(indexPath) })
	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && bot.Global.AcceptsWebhookPath(c.Request.URL.Path) {
			bot.Global.WebhookHandler()(c.Writer, c.Request)
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		if isBackendPath(c.Request.URL.Path, cfg.AdminRoutePrefix) {
			c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
			return
		}
		if serveDistFile(c, dist) {
			return
		}
		c.File(indexPath)
	})
}

func isBackendPath(path, adminPrefix string) bool {
	if path == adminPrefix || strings.HasPrefix(path, adminPrefix+"/") {
		return true
	}
	if path == "/telegram" || strings.HasPrefix(path, "/telegram/") {
		return true
	}
	return path == "/health" || path == "/app-config.js"
}

func serveDistFile(c *gin.Context, dist string) bool {
	requestPath := strings.TrimPrefix(filepath.Clean(c.Request.URL.Path), string(filepath.Separator))
	if requestPath == "." || strings.HasPrefix(requestPath, "..") {
		return false
	}
	fullPath := filepath.Join(dist, requestPath)
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return false
	}
	c.File(fullPath)
	return true
}
