package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"sync"
	"time"
)

var loginCrypto = newChallengeBox()

type challengeBox struct {
	mu         sync.Mutex
	privateKey *rsa.PrivateKey
	nonces     map[string]time.Time
}

func newChallengeBox() *challengeBox {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	return &challengeBox{privateKey: key, nonces: make(map[string]time.Time)}
}

func NewLoginChallenge() (string, string, error) {
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", "", err
	}
	nonce := base64.RawURLEncoding.EncodeToString(nonceBytes)
	loginCrypto.mu.Lock()
	defer loginCrypto.mu.Unlock()
	loginCrypto.cleanupLocked(time.Now())
	loginCrypto.nonces[nonce] = time.Now().Add(3 * time.Minute)
	return nonce, loginCrypto.publicPEM(), nil
}

func DecryptLoginPayload(ciphertextB64, nonce string, out any) error {
	loginCrypto.mu.Lock()
	expiresAt, ok := loginCrypto.nonces[nonce]
	if ok {
		delete(loginCrypto.nonces, nonce)
	}
	loginCrypto.cleanupLocked(time.Now())
	privateKey := loginCrypto.privateKey
	loginCrypto.mu.Unlock()
	if !ok || time.Now().After(expiresAt) {
		return errors.New("login challenge expired")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return err
	}
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, []byte(nonce))
	if err != nil {
		return err
	}
	return json.Unmarshal(plaintext, out)
}

func (b *challengeBox) publicPEM() string {
	der, _ := x509.MarshalPKIXPublicKey(&b.privateKey.PublicKey)
	return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}))
}

func (b *challengeBox) cleanupLocked(now time.Time) {
	for nonce, expiresAt := range b.nonces {
		if now.After(expiresAt) {
			delete(b.nonces, nonce)
		}
	}
}

func HashPassword(password string) string {
	digest := sha256.Sum256([]byte(password))
	return hex.EncodeToString(digest[:])
}
