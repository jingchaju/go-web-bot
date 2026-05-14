package handlers

import (
	"net/http"
	"strings"
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

type encryptedLoginReq struct {
	Nonce   string `json:"nonce"`
	Payload string `json:"payload"`
}

type loginPayload struct {
	Account     string `json:"account"`
	Password    string `json:"password"`
	Fingerprint string `json:"fingerprint"`
	Nonce       string `json:"nonce"`
}

type updateAccountReq struct {
	Account string `json:"account"`
}

type updatePasswordReq struct {
	Password string `json:"password"`
}

func (h AdminHandler) AuthChallenge(c *gin.Context) {
	nonce, publicKey, err := auth.NewLoginChallenge()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "challenge error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"nonce": nonce, "public_key": publicKey})
}

func (h AdminHandler) Login(c *gin.Context) {
	var req encryptedLoginReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Nonce == "" || req.Payload == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	var payload loginPayload
	if err := auth.DecryptLoginPayload(req.Payload, req.Nonce, &payload); err != nil || payload.Nonce != req.Nonce {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "登录挑战已过期，请重试"})
		return
	}
	admin, err := h.DAO.RequireFirst()
	if err != nil || admin.AdminAccount != payload.Account {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "账号或密码错误"})
		return
	}
	if !auth.EqualSignature(payload.Password, admin.PasswordSignature, h.Config.AdminSecret) {
		if !auth.EqualLegacySignature(payload.Password, admin.PasswordSignature, h.Config.AdminSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "账号或密码错误"})
			return
		}
		admin.PasswordSignature = auth.SignPassword(auth.HashPassword(payload.Password), h.Config.AdminSecret)
		_ = h.DAO.Save(admin)
	}
	token, err := auth.IssueJWT(admin.ID, admin.AdminAccount, payload.Fingerprint, h.Config.JWTSecret, time.Duration(h.Config.AdminSessionTTLMinutes)*time.Minute)
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
	if err := bot.Global.Start(*admin, h.Config); err != nil {
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

func (h AdminHandler) UpdateAccount(c *gin.Context) {
	var req updateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	req.Account = strings.TrimSpace(req.Account)
	if len(req.Account) < 4 || len(req.Account) > 32 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "管理员账号长度需为 4-32 位"})
		return
	}
	admin, err := h.DAO.RequireFirst()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "not initialized"})
		return
	}
	admin.AdminAccount = req.Account
	if err := h.DAO.Save(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "账号已更新，请重新登录"})
}

func (h AdminHandler) UpdatePassword(c *gin.Context) {
	var req updatePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "bad request"})
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 128 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "密码长度需为 8-128 位"})
		return
	}
	admin, err := h.DAO.RequireFirst()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "not initialized"})
		return
	}
	admin.PasswordSignature = auth.SignPassword(auth.HashPassword(req.Password), h.Config.AdminSecret)
	if err := h.DAO.Save(admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "密码已更新，请重新登录"})
}
