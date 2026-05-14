package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go-web-bot/internal/auth"
	"go-web-bot/internal/config"
)

func Security(_ config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "same-origin")
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
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
		if err := verifySignedRequest(c, token, fp); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			return
		}
		c.Set("admin_id", claims.AdminID)
		c.Set("admin_account", claims.Account)
		c.Next()
	}
}

func verifySignedRequest(c *gin.Context, token, fingerprint string) error {
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
		return nil
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Errorf("invalid request body")
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	timestamp := c.GetHeader("X-Request-Timestamp")
	sig := c.GetHeader("X-Request-Signature")
	if timestamp == "" || sig == "" {
		return fmt.Errorf("missing request signature")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || time.Since(time.Unix(ts, 0)) > 5*time.Minute || time.Until(time.Unix(ts, 0)) > time.Minute {
		return fmt.Errorf("expired request signature")
	}
	payload := strings.Join([]string{c.Request.Method, c.Request.URL.Path, timestamp, fingerprint, string(body)}, "\n")
	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(payload))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("invalid request signature")
	}
	return nil
}
