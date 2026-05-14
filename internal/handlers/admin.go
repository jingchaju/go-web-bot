package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/auth"
	"go-web-bot/internal/bot"
	"go-web-bot/internal/config"
	"go-web-bot/internal/dao"
	"go-web-bot/internal/models"
)

type AdminHandler struct {
	Config config.Config
	DAO    dao.AdminDAO
}
type loginReq struct {
	Account     string `json:"account"`
	Password    string `json:"password"`
	Fingerprint string `json:"fingerprint"`
}

func (h AdminHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	admin, err := h.DAO.RequireFirst()
	if err != nil || admin.AdminAccount != req.Account || !auth.EqualSignature(req.Password, admin.PasswordSignature, h.Config.AdminSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "账号或密码错误"})
		return
	}
	token, err := auth.IssueJWT(admin.ID, admin.AdminAccount, req.Fingerprint, h.Config.JWTSecret, time.Duration(h.Config.AdminSessionTTLMinutes)*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "token error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token, "account": admin.AdminAccount, "expires_in": h.Config.AdminSessionTTLMinutes * 60})
}
func (h AdminHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"account": c.GetString("admin_account")})
}
func (h AdminHandler) Dashboard(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"users": 0, "bot": bot.Global.Stats(), "server_time": time.Now()})
}
func (h AdminHandler) GetBotConfig(c *gin.Context) {
	admin, err := h.DAO.RequireFirst()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "not initialized"})
		return
	}
	c.JSON(http.StatusOK, admin)
}
func (h AdminHandler) SaveBotConfig(c *gin.Context) {
	var req models.AdminConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	admin, err := h.DAO.RequireFirst()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "not initialized"})
		return
	}
	admin.BotToken, admin.WebhookPort, admin.WebhookPath, admin.WebhookSecret = req.BotToken, req.WebhookPort, req.WebhookPath, req.WebhookSecret
	admin.CommandsJSON, admin.KeyboardJSON, admin.WelcomeHTML = req.CommandsJSON, req.KeyboardJSON, req.WelcomeHTML
	if err := h.DAO.Save(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "save failed"})
		return
	}
	c.JSON(http.StatusOK, admin)
}
func (h AdminHandler) StartBot(c *gin.Context) {
	admin, err := h.DAO.RequireFirst()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "not initialized"})
		return
	}
	if err := bot.Global.Start(*admin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	admin.BotRunning = true
	_ = h.DAO.Save(admin)
	c.JSON(http.StatusOK, bot.Global.Stats())
}
func (h AdminHandler) StopBot(c *gin.Context) {
	bot.Global.Stop()
	admin, _ := h.DAO.RequireFirst()
	if admin != nil {
		admin.BotRunning = false
		_ = h.DAO.Save(admin)
	}
	c.JSON(http.StatusOK, bot.Global.Stats())
}
