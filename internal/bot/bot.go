package bot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"go-web-bot/internal/logger"
	"go-web-bot/internal/models"
	"go-web-bot/internal/pool"
	"go-web-bot/internal/redisclient"
	"gopkg.in/telebot.v4"
)

type Manager struct {
	mu        sync.RWMutex
	Bot       *telebot.Bot
	Config    models.AdminConfig
	Queue     redisclient.Queue
	StartedAt time.Time
	Processed uint64
}

var Global = &Manager{Queue: redisclient.RegisterQueue("telegram_updates", 10*time.Minute)}

type QueuedUpdate struct {
	Update     telebot.Update `json:"update"`
	ReceivedAt time.Time      `json:"received_at"`
}

func (m *Manager) Start(cfg models.AdminConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg.BotToken == "" || cfg.WebhookSecret == "" || cfg.WebhookPath == "" {
		return errors.New("bot token, webhook path and secret are required")
	}
	b, err := telebot.NewBot(telebot.Settings{Token: cfg.BotToken, ParseMode: telebot.ModeHTML, Poller: &telebot.Webhook{Endpoint: &telebot.WebhookEndpoint{PublicURL: cfg.WebhookPath}, SecretToken: cfg.WebhookSecret}})
	if err != nil {
		return err
	}
	m.Bot, m.Config, m.StartedAt = b, cfg, time.Now()
	m.registerHandlers()
	go m.consume(context.Background())
	logger.Info("telegram bot manager started\n")
	return nil
}
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Bot != nil {
		m.Bot.Stop()
	}
	m.Bot = nil
}
func (m *Manager) Running() bool { m.mu.RLock(); defer m.mu.RUnlock(); return m.Bot != nil }
func (m *Manager) registerHandlers() {
	m.Bot.Handle("/start", func(c telebot.Context) error {
		return c.Send(m.Config.WelcomeHTML, &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true})
	})
	m.Bot.Handle(telebot.OnText, func(c telebot.Context) error { logger.Info("text from %d\n", c.Sender().ID); return nil })
	m.Bot.Handle(telebot.OnCallback, func(c telebot.Context) error { logger.Info("callback %s\n", c.Callback().Data); return c.Respond() })
}
func (m *Manager) Enqueue(update telebot.Update) error {
	raw, _ := json.Marshal(QueuedUpdate{Update: update, ReceivedAt: time.Now()})
	return m.Queue.Push(context.Background(), string(raw))
}
func (m *Manager) consume(ctx context.Context) {
	for m.Running() {
		payload, err := m.Queue.Pop(ctx, time.Second)
		if err != nil {
			continue
		}
		_ = pool.Submit(func() { m.dispatch(payload) })
	}
}
func (m *Manager) dispatch(payload string) {
	var qu QueuedUpdate
	if err := json.Unmarshal([]byte(payload), &qu); err != nil {
		return
	}
	if time.Since(qu.ReceivedAt) > 10*time.Minute {
		logger.Warning("discarded expired telegram update\n")
		return
	}
	m.mu.RLock()
	b := m.Bot
	m.mu.RUnlock()
	if b == nil {
		return
	}
	b.ProcessUpdate(qu.Update)
	m.Processed++
}
func (m *Manager) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		secret := m.Config.WebhookSecret
		m.mu.RUnlock()
		if secret == "" || r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != secret {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		var update telebot.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if err := m.Enqueue(update); err != nil {
			http.Error(w, "queue error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}
}
func (m *Manager) Stats() map[string]any {
	return map[string]any{"running": m.Running(), "started_at": m.StartedAt, "processed": m.Processed}
}
