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

	// 1. 基础监控与全局动态配置
	r.GET("/health", func(c *gin.Context) { 
		c.JSON(http.StatusOK, gin.H{"status": "ok"}) 
	})
    
	// 引用外部定义（确保 internal/router/app_config.go 中有此函数）
	r.GET("/app-config.js", appConfigHandler(cfg))

	adminDAO := dao.NewAdminDAO(db)
	h := handlers.AdminHandler{Config: cfg, DAO: adminDAO}

	// 2. 管理后台 API 分组
	// 使用自定义前缀 + 安全过滤（防XSS、CORS、请求签名校验）
	api := r.Group(cfg.AdminRoutePrefix, middleware.Security(cfg))
	{
		// 登录流程：挑战码获取 -> RSA加密登录
		api.GET("/auth/challenge", h.AuthChallenge)
		api.POST("/login", h.Login)

		// 3. 需要 JWT 验证的私有接口分组
		// 增强：强制鉴权，且所有 API 交互在 middleware 中完成解密/验签
		private := api.Group("", middleware.JWT(cfg))
		{
			private.GET("/me", h.Me)
			private.GET("/dashboard", h.Dashboard)
			
			// Bot 管理
			private.GET("/bot/config", h.GetBotConfig)
			private.POST("/bot/config", h.SaveBotConfig) // 成功后前端需弹窗交互
			private.POST("/bot/start", h.StartBot)       // 成功后前端触发通知
			private.POST("/bot/stop", h.StopBot)
			
			// 系统设置：更改用户名/密码
			// 逻辑说明：成功后后端应清除 Token，前端拦截返回码执行登出
			private.POST("/settings/account", h.UpdateAccount)
			private.POST("/settings/password", h.UpdatePassword)
		}
	}

	// 4. Telegram Webhook 接口
	// 注意：确保路径与后台配置一致。建议使用动态路径：/telegram/:token
	r.POST("/telegram/webhook", gin.WrapF(bot.Global.WebhookHandler()))

	// 5. 注册前端静态文件服务
	// 引用外部定义（确保 internal/router/frontend.go 中有此函数）
	registerFrontend(r, cfg)

	return r
}


