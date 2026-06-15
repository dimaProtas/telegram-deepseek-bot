package main

import (
	"os"
	"os/signal"
	"syscall"

	"telegram-deepseek-bot/internal/config"
	"telegram-deepseek-bot/internal/deepseek"
	"telegram-deepseek-bot/internal/logger"
	"telegram-deepseek-bot/internal/opencode"
	"telegram-deepseek-bot/internal/storage"
	"telegram-deepseek-bot/internal/telegram"
)

func main() {
	println("DEBUG: loading config...")
	cfg := config.LoadConfig()
	println("DEBUG: config loaded, token len:", len(cfg.TelegramToken), "proxy:", cfg.ProxyConf.Addr)
	l := logger.NewLogger(cfg.LogLevel)

	if cfg.TelegramToken == "" || cfg.DeepSeekAPIKey == "" {
		l.Error("Missing required environment variables: TELEGRAM_BOT_TOKEN or DEEPSEEK_API_KEY")
		os.Exit(1)
	}

	println("DEBUG: creating deepseek client...")
	dsClient := deepseek.NewClient(cfg.DeepSeekAPIKey, cfg.DeepSeekAPIURL, cfg.ProxyConf.Addr, cfg.ProxyConf.Username, cfg.ProxyConf.Password)
	println("DEBUG: deepseek client created")
	store := storage.NewMemoryStorage(cfg.ConversationTTL)
	agent := opencode.NewAgent(cfg)

	println("DEBUG: creating bot...")
	bot, err := telegram.NewBot(cfg, dsClient, store, l, agent)
	if err != nil {
		l.Error("Failed to initialize bot: " + err.Error())
		os.Exit(1)
	}
	println("DEBUG: bot created")

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		bot.Start()
	}()

	l.Info("Bot is running. Press Ctrl+C to stop.")
	<-stop
	l.Info("Shutting down bot...")
}
