package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TelegramID int64     `gorm:"uniqueIndex" json:"telegram_id"`
	Username   string    `gorm:"size:128" json:"username"`
	FirstName  string    `gorm:"size:128" json:"first_name"`
	LastName   string    `gorm:"size:128" json:"last_name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type AdminConfig struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	AdminAccount      string    `gorm:"size:6;uniqueIndex" json:"admin_account"`
	PasswordSignature string    `gorm:"size:512" json:"-"`
	BotToken          string    `gorm:"size:256" json:"bot_token"`
	WebhookPort       string    `gorm:"size:16" json:"webhook_port"`
	WebhookPath       string    `gorm:"size:128" json:"webhook_path"`
	WebhookSecret     string    `gorm:"size:256" json:"webhook_secret"`
	CommandsJSON      string    `gorm:"type:text" json:"commands_json"`
	KeyboardJSON      string    `gorm:"type:text" json:"keyboard_json"`
	WelcomeHTML       string    `gorm:"type:text" json:"welcome_html"`
	BotRunning        bool      `json:"bot_running"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
