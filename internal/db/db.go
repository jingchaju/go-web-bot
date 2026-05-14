package db

import (
	"database/sql"
	"time"

	"go-web-bot/internal/config"
	"go-web-bot/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var Conn *gorm.DB

func Init(cfg config.Config) error {
	d, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{NowFunc: func() time.Time { return time.Now().In(time.FixedZone("Asia/Shanghai", 8*60*60)) }, NamingStrategy: schema.NamingStrategy{SingularTable: true}})
	if err != nil {
		return err
	}
	Conn = d
	sqlDB, err := Conn.DB()
	if err != nil {
		return err
	}
	configurePool(sqlDB, cfg)
	return Conn.AutoMigrate(&models.User{}, &models.AdminConfig{})
}
func configurePool(sqlDB *sql.DB, cfg config.Config) {
	sqlDB.SetMaxIdleConns(cfg.PostgresMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.PostgresMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute)
}
func Get() *gorm.DB { return Conn }
