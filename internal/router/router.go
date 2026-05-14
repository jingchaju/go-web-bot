package router

import (
	"encoding/json"
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

	// 基础监控与配置
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/app-config.js", appConfigHandler(cfg))

	adminDAO := dao.NewAdminDAO(db)
	h := handlers.AdminHandler{Config: cfg, DAO: adminDAO}

	// 管理后台 API 分组 (RSA 登录与请求签名验证)
	api := r.Group(cfg.AdminRoutePrefix, middleware.Security(cfg))
	{
		api.GET("/auth/challenge", h.AuthChallenge)
		api.POST("/login", h.Login)

		// 需要 JWT 验证的私有接口
		private := api.Group("", middleware.JWT(cfg))
		{
			private.GET("/me", h.Me)
			private.GET("/dashboard", h.Dashboard)
			private.GET("/bot/config", h.GetBotConfig)
			private.POST("/bot/config", h.SaveBotConfig)
			private.POST("/bot/start", h.StartBot)
			private.POST("/bot/stop", h.StopBot)
			private.POST("/settings/account", h.UpdateAccount)
			private.POST("/settings/password", h.UpdatePassword)
		}
	}

	// Telegram Webhook 接口
	r.POST("/telegram/webhook", gin.WrapF(bot.Global.WebhookHandler()))

	// 注册前端静态文件服务
	registerFrontend(r, cfg)

	return r
}

// 动态生成前端配置，解决登录接口 404 的关键
func appConfigHandler(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		payload, _ := json.Marshal(gin.H{
			"adminRoutePrefix": cfg.AdminRoutePrefix,
		})
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		c.String(http.StatusOK, "window.__APP_CONFIG__ = %s;", payload)
	}
}

func registerFrontend(r *gin.Engine, cfg config.Config) {
	dist := cfg.FrontendDist
	indexPath := filepath.Join(dist, "index.html")

	// 检查前端静态目录是否存在
	if _, err := os.Stat(indexPath); err != nil {
		r.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "telegram bot admin service is running") })
		return
	}

	r.GET("/", func(c *gin.Context) { c.File(indexPath) })

	// 处理单页应用 (SPA) 路由冲突
	r.NoRoute(func(c *gin.Context) {
		// 1. 补偿处理 Webhook 请求 (针对某些 Nginx 配置转发导致的路径偏差)
		if c.Request.Method == http.MethodPost && strings.HasPrefix(c.Request.URL.Path, "/telegram") {
			bot.Global.WebhookHandler()(c.Writer, c.Request)
			return
		}

		// 2. 如果是后端 API 路径但未匹配到，直接返回 404 而不重定向到 index.html
		if isBackendPath(c.Request.URL.Path, cfg.AdminRoutePrefix) {
			c.JSON(http.StatusNotFound, gin.H{"message": "API endpoint not found"})
			return
		}

		// 3. 尝试读取静态资源 (js/css/images)
		if serveDistFile(c, dist) {
			return
		}

		// 4. 其他所有 GET 请求重定向到 index.html，交给前端 Vue/React Router 处理
		if c.Request.Method == http.MethodGet {
			c.File(indexPath)
			return
		}

		c.JSON(http.StatusNotFound, gin.H{"message": "not found"})
	})
}

// 判定是否属于后端保留路径
func isBackendPath(path, adminPrefix string) bool {
	if path == adminPrefix || strings.HasPrefix(path, adminPrefix+"/") {
		return true
	}
	if strings.HasPrefix(path, "/telegram") {
		return true
	}
	// 排除基础服务路径
	reserved := []string{"/health", "/app-config.js"}
	for _, p := range reserved {
		if path == p {
			return true
		}
	}
	return false
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
