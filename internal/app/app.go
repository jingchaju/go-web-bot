package app

import (
	"go-web-bot/internal/config"
	"go-web-bot/internal/dao"
	"go-web-bot/internal/db"
	"go-web-bot/internal/httpclient"
	"go-web-bot/internal/logger"
	"go-web-bot/internal/pool"
	"go-web-bot/internal/redisclient"
	"go-web-bot/internal/router"
)

func Run() {
	cfg := config.Load()
	logger.Init(cfg.LoggerLevel, cfg.LoggerToFile, cfg.LoggerFile)
	if err := pool.Init(cfg.AntsPoolSize); err != nil {
		logger.Fatal("ants pool init failed: %v\n", err)
	}
	defer pool.Release()
	httpclient.Init(cfg.AntsPoolSize)
	if err := redisclient.Init(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB, cfg.RedisPoolSize); err != nil {
		logger.Fatal("redis init failed: %v\n", err)
	}
	if err := db.Init(cfg); err != nil {
		logger.Fatal("postgres init failed: %v\n", err)
	}
	if err := dao.NewAdminDAO(db.Get()).EnsureDefault(cfg); err != nil {
		logger.Fatal("admin init failed: %v\n", err)
	}
	if err := router.New(cfg, db.Get()).Run(cfg.HTTPAddr); err != nil {
		logger.Fatal("server stopped: %v\n", err)
	}
}
