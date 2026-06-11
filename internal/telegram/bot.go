package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"telegram-deepseek-bot/internal/config"
	"telegram-deepseek-bot/internal/deepseek"
	"telegram-deepseek-bot/internal/logger"
	"telegram-deepseek-bot/internal/opencode"
	"telegram-deepseek-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"golang.org/x/net/proxy"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	cfg    *config.Config
	ds     *deepseek.Client
	store  *storage.MemoryStorage
	logger *logger.Logger
	agent  *opencode.Agent
}

func NewBot(cfg *config.Config, ds *deepseek.Client, store *storage.MemoryStorage, log *logger.Logger, agent *opencode.Agent) (*Bot, error) {
	var bot *tgbotapi.BotAPI
	var err error

	if cfg.ProxyConf.Addr != "" {
		httpClient, err := createProxyHTTPClient(cfg.ProxyConf.Addr, cfg.ProxyConf.Username, cfg.ProxyConf.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to create proxy HTTP client: %w", err)
		}
		bot, err = tgbotapi.NewBotAPIWithClient(cfg.TelegramToken, tgbotapi.APIEndpoint, httpClient)
	} else {
		bot, err = tgbotapi.NewBotAPI(cfg.TelegramToken)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %w", err)
	}

	return &Bot{
		api:    bot,
		cfg:    cfg,
		ds:     ds,
		store:  store,
		logger: log,
		agent:  agent,
	}, nil
}

func (b *Bot) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)

	b.logger.Info("Bot started and polling updates...")

	for update := range updates {
		if update.Message == nil {
			continue
		}

		if !b.isAllowed(update.Message) {
			b.logger.Warn(fmt.Sprintf("Unauthorized user %d tried to access bot", update.Message.From.ID))
			continue
		}

		b.handleMessage(update.Message)
	}
}

func (b *Bot) isAllowed(msg *tgbotapi.Message) bool {
	if len(b.cfg.AllowedUserIDs) == 0 {
		return true
	}

	for _, id := range b.cfg.AllowedUserIDs {
		if id == msg.From.ID {
			return true
		}
	}
	return false
}

func (b *Bot) handleMessage(msg *tgbotapi.Message) {
	text := msg.Text
	chatID := msg.Chat.ID

	if strings.HasPrefix(text, "/start") {
		b.sendHelp(chatID)
		return
	}

	if strings.HasPrefix(text, "/help") {
		b.sendHelp(chatID)
		return
	}

	if strings.HasPrefix(text, "/clear") {
		b.store.Clear(chatID)
		b.sendMessage(chatID, "История диалога очищена.")
		return
	}

	if strings.HasPrefix(text, "/state") {
		b.handleState(chatID)
		return
	}

	if strings.HasPrefix(text, "/retry") {
		b.handleRetry(chatID)
		return
	}

	if strings.HasPrefix(text, "/mode") {
		b.handleMode(chatID, text)
		return
	}

	if strings.HasPrefix(text, "/run ") {
		prompt := strings.TrimPrefix(text, "/run ")
		b.handleRun(chatID, msg.From.ID, prompt)
		return
	}

	if strings.HasPrefix(text, "/code ") {
		prompt := strings.TrimPrefix(text, "/code ")
		b.handleCode(chatID, prompt)
		return
	}

	if strings.HasPrefix(text, "/review ") {
		file := strings.TrimPrefix(text, "/review ")
		b.handleFileCommand(chatID, file, "review")
		return
	}

	if strings.HasPrefix(text, "/explain ") {
		file := strings.TrimPrefix(text, "/explain ")
		b.handleFileCommand(chatID, file, "explain")
		return
	}

	if strings.HasPrefix(text, "/test ") {
		file := strings.TrimPrefix(text, "/test ")
		b.handleFileCommand(chatID, file, "test")
		return
	}

	if strings.HasPrefix(text, "/docs ") {
		file := strings.TrimPrefix(text, "/docs ")
		b.handleFileCommand(chatID, file, "docs")
		return
	}

	// Basic chat command: /chat <prompt>
	if strings.HasPrefix(text, "/chat ") {
		prompt := strings.TrimPrefix(text, "/chat ")
		b.handleChat(chatID, prompt)
		return
	}

	// Default: treat as chat if no command
	b.handleChat(chatID, text)
}

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	b.api.Send(msg)
}

func (b *Bot) handleRun(chatID int64, userID int64, prompt string) {
	b.logger.Info(fmt.Sprintf("[RUN] Received at %s, userID %d, chatID %d, prompt: %s",
		time.Now().Format(time.RFC3339), userID, chatID, prompt))
	b.sendMessage(chatID, "⏳ Запуск OpenCode Agent...")

	go func() {
		startTime := time.Now()
		b.logger.Info(fmt.Sprintf("[RUN] Starting OpenCode Agent for userID %d, chatID %d at %s",
			userID, chatID, startTime.Format(time.RFC3339)))

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(b.cfg.OpenCodeTimeout)*time.Second)
		defer cancel()

		result, err := b.agent.Execute(ctx, prompt)
		elapsed := time.Since(startTime)

		if err != nil {
			b.logger.Error(fmt.Sprintf("[RUN] OpenCode Agent error for userID %d, chatID %d (elapsed %s): %v",
				userID, chatID, elapsed, err))
			b.sendMessage(chatID, "Ошибка выполнения задачи. Попробуйте повторить запрос позже.")
		} else {
			b.logger.Info(fmt.Sprintf("[RUN] OpenCode Agent completed for userID %d, chatID %d (elapsed %s), status: success",
				userID, chatID, elapsed))
			if result == "" {
				b.sendMessage(chatID, "✅ Задача выполнена успешно (ответ пуст)")
			} else {
				b.sendLongMessage(chatID, result)
			}
		}
	}()
}

func (b *Bot) sendLongMessage(chatID int64, text string) {
	const maxLen = 4096
	if len(text) <= maxLen {
		b.sendMessage(chatID, text)
		return
	}

	parts := splitMessage(text, maxLen)
	for _, part := range parts {
		b.sendMessage(chatID, part)
	}
}

func splitMessage(text string, maxLen int) []string {
	var parts []string
	runes := []rune(text)

	for i := 0; i < len(runes); i += maxLen {
		end := i + maxLen
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[i:end]))
	}
	return parts
}

func createProxyHTTPClient(addr, username, password string) (*http.Client, error) {
	var auth *proxy.Auth
	if username != "" || password != "" {
		auth = &proxy.Auth{User: username, Password: password}
	}

	dialer, err := proxy.SOCKS5("tcp", addr, auth, proxy.Direct)
	if err != nil {
		return nil, err
	}

	return &http.Client{
		Transport: &http.Transport{Dial: dialer.Dial},
		Timeout:   60 * time.Second,
	}, nil
}

func (b *Bot) sendHelp(chatID int64) {
	helpText := `Доступные команды:
/chat <текст> - Запрос к AI
/run <текст> - Локальное исполнение задачи через OpenCode Agent
/code <текст> - Генерация кода (использует coder модель)
/mode <модель> - Сменить модель (chat, coder, reasoner)
/review <файл> - Ревью кода
/explain <файл> - Объяснение кода
/test <файл> - Генерация тестов
/docs <файл> - Документация
/state - Статистика токенов
/retry - Повторить последний запрос
/clear - Очистить историю
/help - Справка`
	b.sendMessage(chatID, helpText)
}
