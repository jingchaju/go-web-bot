package dao

import (
	"errors"

	"go-web-bot/internal/auth"
	"go-web-bot/internal/config"
	"go-web-bot/internal/logger"
	"go-web-bot/internal/models"
	"gorm.io/gorm"
)

type AdminDAO struct{ DB *gorm.DB }

func NewAdminDAO(db *gorm.DB) AdminDAO { return AdminDAO{DB: db} }
func (d AdminDAO) First() (*models.AdminConfig, error) {
	var a models.AdminConfig
	err := d.DB.First(&a).Error
	return &a, err
}
func (d AdminDAO) Save(a *models.AdminConfig) error { return d.DB.Save(a).Error }
func (d AdminDAO) EnsureDefault(cfg config.Config) error {
	var count int64
	if err := d.DB.Model(&models.AdminConfig{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	account, password, sig, err := auth.GenerateAdminCredential(cfg.AdminSecret)
	if err != nil {
		return err
	}
	admin := &models.AdminConfig{AdminAccount: account, PasswordSignature: sig, WebhookPath: "/telegram/webhook", WebhookPort: "8080", CommandsJSON: "[]", KeyboardJSON: "[]", WelcomeHTML: "欢迎使用 Telegram Bot"}
	if err := d.DB.Create(admin).Error; err != nil {
		return err
	}
	logger.Warning("INITIAL ADMIN GENERATED account=%s password=%s save it immediately\n", account, password)
	return nil
}
func (d AdminDAO) RequireFirst() (*models.AdminConfig, error) {
	a, err := d.First()
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return a, err
}
