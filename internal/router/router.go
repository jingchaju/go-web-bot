package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/bot"
	"go-web-bot/internal/config"
	"go-web-bot/internal/dao"
	"go-web-bot/internal/handlers"
	"go-web-bot/internal/middleware"
	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	adminDAO := dao.NewAdminDAO(db)
	h := handlers.AdminHandler{Config: cfg, DAO: adminDAO}
	api := r.Group(cfg.AdminRoutePrefix, middleware.Security(cfg))
	api.POST("/login", h.Login)
	private := api.Group("", middleware.JWT(cfg))
	private.GET("/me", h.Me)
	private.GET("/dashboard", h.Dashboard)
	private.GET("/bot/config", h.GetBotConfig)
	private.POST("/bot/config", h.SaveBotConfig)
	private.POST("/bot/start", h.StartBot)
	private.POST("/bot/stop", h.StopBot)
	r.POST("/telegram/webhook", gin.WrapF(bot.Global.WebhookHandler()))
	registerFrontend(r, cfg)
	return r
}

func registerFrontend(r *gin.Engine, cfg config.Config) {
	dist := cfg.FrontendDist
	indexPath := filepath.Join(dist, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "telegram bot admin service") })
		return
	}

	r.GET("/", func(c *gin.Context) { c.File(indexPath) })
	r.NoRoute(func(c *gin.Context) {
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
	return strings.HasPrefix(path, adminPrefix) || strings.HasPrefix(path, "/telegram/") || path == "/health"
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
