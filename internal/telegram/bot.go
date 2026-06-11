package telegram

import (
	"context"
	"fmt"
	"net/http"
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
	router *CommandRouter
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

	b := &Bot{
		api:    bot,
		cfg:    cfg,
		ds:     ds,
		store:  store,
		logger: log,
		agent:  agent,
		router: NewCommandRouter(),
	}
	b.registerCommands()
	return b, nil
}

func (b *Bot) registerCommands() {
	registerAgentCommands(b.router)

	defaultAgent := b.cfg.OpenCodeDefaultAgent
	if defaultAgent != "" {
		b.router.Register(NewAgentCommand("/run", defaultAgent, "OpenCode Agent"))
	} else {
		b.router.Register(&prefixCommand{
			name:   "/run",
			prefix: "/run",
			fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
				if arg == "" {
					bot.sendMessage(msg.Chat.ID, "Укажите задачу после команды. Пример: /run <задача>")
					return
				}
				bot.handleRun(msg.Chat.ID, msg.From.ID, arg)
			},
		})
	}

	b.router.Register(&exactCommand{name: "/start", prefix: "/start", fn: func(bot *Bot, msg *tgbotapi.Message) { bot.sendHelp(msg.Chat.ID) }})
	b.router.Register(&exactCommand{name: "/help", prefix: "/help", fn: func(bot *Bot, msg *tgbotapi.Message) { bot.sendHelp(msg.Chat.ID) }})

	b.router.Register(&exactCommand{name: "/clear", prefix: "/clear", fn: func(bot *Bot, msg *tgbotapi.Message) {
		bot.store.Clear(msg.Chat.ID)
		bot.sendMessage(msg.Chat.ID, "История диалога очищена.")
	}})

	b.router.Register(&exactCommand{name: "/state", prefix: "/state", fn: func(bot *Bot, msg *tgbotapi.Message) { bot.handleState(msg.Chat.ID) }})
	b.router.Register(&exactCommand{name: "/retry", prefix: "/retry", fn: func(bot *Bot, msg *tgbotapi.Message) { bot.handleRetry(msg.Chat.ID) }})

	b.router.Register(&prefixCommand{name: "/mode", prefix: "/mode", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleMode(msg.Chat.ID, msg.Text)
	}})

	b.router.Register(&prefixCommand{name: "/code", prefix: "/code", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleCode(msg.Chat.ID, arg)
	}})

	b.router.Register(&prefixCommand{name: "/review", prefix: "/review", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleFileCommand(msg.Chat.ID, arg, "review")
	}})

	b.router.Register(&prefixCommand{name: "/explain", prefix: "/explain", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleFileCommand(msg.Chat.ID, arg, "explain")
	}})

	b.router.Register(&prefixCommand{name: "/test", prefix: "/test", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleFileCommand(msg.Chat.ID, arg, "test")
	}})

	b.router.Register(&prefixCommand{name: "/docs", prefix: "/docs", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleFileCommand(msg.Chat.ID, arg, "docs")
	}})

	b.router.Register(&prefixCommand{name: "/chat", prefix: "/chat", fn: func(bot *Bot, msg *tgbotapi.Message, arg string) {
		bot.handleChat(msg.Chat.ID, arg)
	}})

	b.router.Register(&fallbackCommand{name: "chat (default)", fn: func(bot *Bot, msg *tgbotapi.Message) {
		bot.handleChat(msg.Chat.ID, msg.Text)
	}})
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

		b.router.Dispatch(b, update.Message)
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
	helpText := `🤖 OpenCode агенты (выполнение задач):
/run <задача> — общий агент (по умолчанию)
/explore <задача> — исследование кодовой базы
/go-senior <задача> — написание Go-кода (Senior уровень)
/go-review <промпт> — ревью Go-кода
/react-dev <задача> — разработка React-компонентов
/react-review <промпт> — ревью React-кода
/write-tests <промпт> — генерация тестов
/tz <задача> — создание технического задания

💬 Чат с DeepSeek:
/chat <текст> — запрос к AI
/code <текст> — генерация кода (coder модель)
/mode <модель> — смена модели (chat, coder, reasoner)

📄 Работа с файлами:
/review <файл> — ревью кода
/explain <файл> — объяснение кода
/test <файл> — генерация тестов
/docs <файл> — документация

⚙️ Управление:
/state — статистика токенов
/retry — повторить последний запрос
/clear — очистить историю
/help — справка`
	b.sendMessage(chatID, helpText)
}
