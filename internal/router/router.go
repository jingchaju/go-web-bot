package router

import (
	"net/http"

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
	r.GET("/app-config.js", runtimeAppConfigHandler(cfg))
	adminDAO := dao.NewAdminDAO(db)
	h := handlers.AdminHandler{Config: cfg, DAO: adminDAO}
	api := r.Group(cfg.AdminRoutePrefix, middleware.Security(cfg))
	api.GET("/auth/challenge", h.AuthChallenge)
	api.POST("/login", h.Login)
	private := api.Group("", middleware.JWT(cfg))
	private.GET("/me", h.Me)
	private.GET("/dashboard", h.Dashboard)
	private.GET("/bot/config", h.GetBotConfig)
	private.POST("/bot/config", h.SaveBotConfig)
	private.POST("/bot/start", h.StartBot)
	private.POST("/bot/stop", h.StopBot)
	private.POST("/settings/account", h.UpdateAccount)
	private.POST("/settings/password", h.UpdatePassword)
	r.POST("/telegram/webhook", gin.WrapF(bot.Global.WebhookHandler()))
	installFrontendRoutes(r, cfg)
	return r
}
