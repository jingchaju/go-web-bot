package bot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"go-web-bot/internal/config"
	"go-web-bot/internal/logger"
	"go-web-bot/internal/models"
	"go-web-bot/internal/pool"
	"go-web-bot/internal/redisclient"
	"gopkg.in/telebot.v4"
)

type Manager struct {
	mu          sync.RWMutex
	Bot         *telebot.Bot
	Config      models.AdminConfig
	Queue       redisclient.Queue
	StartedAt   time.Time
	Processed   uint64
	WebhookPath string
}

var Global = &Manager{Queue: redisclient.RegisterQueue("telegram_updates", 10*time.Minute)}

type QueuedUpdate struct {
	Update     telebot.Update `json:"update"`
	ReceivedAt time.Time      `json:"received_at"`
}

type configuredCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

type configuredButton struct {
	Text string `json:"text"`
	URL  string `json:"url"`
	Data string `json:"data"`
}

func (m *Manager) Start(cfg models.AdminConfig, appCfg config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if cfg.BotToken == "" || cfg.WebhookSecret == "" || cfg.WebhookPath == "" {
		return errors.New("bot token, webhook path and secret are required")
	}
	publicURL, webhookPath, err := normalizeWebhook(cfg.WebhookPath, appCfg.PublicBaseURL)
	if err != nil {
		return err
	}
	b, err := telebot.NewBot(telebot.Settings{Token: cfg.BotToken, ParseMode: telebot.ModeHTML})
	if err != nil {
		return err
	}
	if err := b.SetWebhook(&telebot.Webhook{Endpoint: &telebot.WebhookEndpoint{PublicURL: publicURL}, SecretToken: cfg.WebhookSecret, DropUpdates: true}); err != nil {
		return fmt.Errorf("telegram setWebhook failed: %w", err)
	}
	if err := applyCommands(b, cfg.CommandsJSON); err != nil {
		return err
	}
	m.Bot, m.Config, m.StartedAt, m.WebhookPath = b, cfg, time.Now(), webhookPath
	m.registerHandlers()
	go m.consume(context.Background())
	logger.Info("telegram bot manager started webhook=%s public_url=%s\n", webhookPath, publicURL)
	return nil
}
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Bot != nil {
		_ = m.Bot.RemoveWebhook(false)
		m.Bot.Stop()
	}
	m.Bot = nil
	m.WebhookPath = ""
}
func (m *Manager) Running() bool { m.mu.RLock(); defer m.mu.RUnlock(); return m.Bot != nil }
func (m *Manager) registerHandlers() {
	m.Bot.Handle("/start", func(c telebot.Context) error {
		return c.Send(m.Config.WelcomeHTML, buildInlineKeyboard(m.Config.KeyboardJSON))
	})
	m.Bot.Handle("/help", func(c telebot.Context) error {
		return c.Send(m.Config.WelcomeHTML, buildInlineKeyboard(m.Config.KeyboardJSON))
	})
	m.Bot.Handle("/contact", func(c telebot.Context) error {
		return c.Send("请通过下方按钮联系客服。", buildInlineKeyboard(m.Config.KeyboardJSON))
	})
	m.Bot.Handle(telebot.OnText, func(c telebot.Context) error {
		logger.Info("text from %d\n", c.Sender().ID)
		return c.Send(m.Config.WelcomeHTML, buildInlineKeyboard(m.Config.KeyboardJSON))
	})
	m.Bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		logger.Info("callback %s\n", c.Callback().Data)
		return c.Respond(&telebot.CallbackResponse{Text: "已收到操作：" + c.Callback().Data})
	})
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
func (m *Manager) AcceptsWebhookPath(path string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Bot != nil && m.WebhookPath != "" && path == m.WebhookPath
}
func (m *Manager) WebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		secret := m.Config.WebhookSecret
		expectedPath := m.WebhookPath
		m.mu.RUnlock()
		if expectedPath != "" && r.URL.Path != expectedPath {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
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
	return map[string]any{"running": m.Running(), "started_at": m.StartedAt, "processed": m.Processed, "webhook_path": m.WebhookPath}
}

func normalizeWebhook(rawPath, publicBaseURL string) (string, string, error) {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "", "", errors.New("webhook path is required")
	}
	if u, err := url.Parse(rawPath); err == nil && u.Scheme != "" && u.Host != "" {
		if u.Scheme != "https" {
			return "", "", errors.New("telegram webhook public URL must use https")
		}
		return rawPath, u.EscapedPath(), nil
	}
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	if publicBaseURL == "" {
		return "", "", errors.New("webhook path is relative; set PUBLIC_BASE_URL=https://your-domain.com in .env or enter a full https webhook URL")
	}
	base, err := url.Parse(strings.TrimRight(publicBaseURL, "/"))
	if err != nil || base.Scheme != "https" || base.Host == "" {
		return "", "", errors.New("PUBLIC_BASE_URL must be a full https URL, for example https://gobot.nmbot.org")
	}
	return base.String() + rawPath, rawPath, nil
}

func applyCommands(b *telebot.Bot, raw string) error {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var configured []configuredCommand
	if err := json.Unmarshal([]byte(raw), &configured); err != nil {
		return fmt.Errorf("commands json invalid: %w", err)
	}
	commands := make([]telebot.Command, 0, len(configured))
	for _, cmd := range configured {
		text := strings.TrimPrefix(strings.TrimSpace(cmd.Command), "/")
		desc := strings.TrimSpace(cmd.Description)
		if text == "" || desc == "" {
			continue
		}
		commands = append(commands, telebot.Command{Text: text, Description: desc})
	}
	if len(commands) == 0 {
		return nil
	}
	if err := b.SetCommands(commands); err != nil {
		return fmt.Errorf("set bot commands failed: %w", err)
	}
	return nil
}

func buildInlineKeyboard(raw string) *telebot.SendOptions {
	if strings.TrimSpace(raw) == "" {
		return &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true}
	}
	var configured []configuredButton
	if err := json.Unmarshal([]byte(raw), &configured); err != nil {
		logger.Warning("keyboard json invalid: %v\n", err)
		return &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true}
	}
	markup := &telebot.ReplyMarkup{}
	buttons := make([]telebot.Btn, 0, len(configured))
	for _, item := range configured {
		text := strings.TrimSpace(item.Text)
		if text == "" {
			continue
		}
		if item.URL != "" {
			buttons = append(buttons, markup.URL(text, strings.TrimSpace(item.URL)))
			continue
		}
		if item.Data != "" {
			buttons = append(buttons, markup.Data(text, strings.TrimSpace(item.Data), strings.TrimSpace(item.Data)))
		}
	}
	if len(buttons) > 0 {
		markup.Inline(markup.Split(3, buttons)...)
	}
	return &telebot.SendOptions{ParseMode: telebot.ModeHTML, DisableWebPagePreview: true, ReplyMarkup: markup}
}
