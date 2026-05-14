package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go-web-bot/internal/ctime"
)

type Claims struct {
	AdminID     uint   `json:"admin_id"`
	Account     string `json:"account"`
	Fingerprint string `json:"fingerprint"`
	jwt.RegisteredClaims
}

func GenerateAdminCredential(adminSecret string) (string, string, string, error) {
	seed := make([]byte, 48)
	if _, err := rand.Read(seed); err != nil {
		return "", "", "", err
	}
	now := ctime.Timestamp()
	digest := sha256.Sum256([]byte(fmt.Sprintf("%d:%s", now, hex.EncodeToString(seed))))
	accountNum := int(digest[0])<<16 | int(digest[1])<<8 | int(digest[2])
	account := fmt.Sprintf("%06d", accountNum%1000000)
	passwordBytes := sha256.Sum256(append(seed, digest[:]...))
	password := hex.EncodeToString(passwordBytes[:])
	return account, password, SignPassword(HashPassword(password), adminSecret), nil
}
func SignPassword(password, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(password))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
func EqualSignature(password, stored, secret string) bool {
	return hmac.Equal([]byte(SignPassword(HashPassword(password), secret)), []byte(stored))
}

func EqualLegacySignature(password, stored, secret string) bool {
	return hmac.Equal([]byte(SignPassword(password, secret)), []byte(stored))
}
func SignPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
func IssueJWT(adminID uint, account, fingerprint, secret string, ttl time.Duration) (string, error) {
	claims := Claims{AdminID: adminID, Account: account, Fingerprint: fingerprint, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)), IssuedAt: jwt.NewNumericDate(time.Now())}}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
func ParseJWT(tokenStr, secret string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
