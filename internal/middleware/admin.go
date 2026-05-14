package middleware

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/auth"
	"go-web-bot/internal/config"
)

func Security(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			body, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			if sig := c.GetHeader("X-Payload-Signature"); sig == "" || sig != auth.SignPayload(body, cfg.JWTSecret) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid payload signature"})
				return
			}
		}
		c.Next()
	}
}
func JWT(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		claims, err := auth.ParseJWT(token, cfg.JWTSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "invalid token"})
			return
		}
		fp := c.GetHeader("X-Device-Fingerprint")
		if claims.Fingerprint != fp {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "fingerprint mismatch"})
			return
		}
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_account", claims.Account)
		c.Next()
	}
}
