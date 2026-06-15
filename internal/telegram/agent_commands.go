package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AgentCommand struct {
	prefix      string
	agentName   string
	displayName string
}

func NewAgentCommand(prefix, agentName, displayName string) *AgentCommand {
	return &AgentCommand{
		prefix:      prefix,
		agentName:   agentName,
		displayName: displayName,
	}
}

func (c *AgentCommand) Name() string {
	return c.displayName
}

func (c *AgentCommand) Matches(text string) bool {
	return strings.HasPrefix(text, c.prefix+" ") || strings.HasPrefix(text, c.prefix+"@")
}

func (c *AgentCommand) Execute(bot *Bot, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID
	userID := msg.From.ID
	prompt := strings.TrimPrefix(msg.Text, c.prefix)
	prompt = strings.TrimPrefix(prompt, "@"+bot.api.Self.UserName)
	prompt = strings.TrimSpace(prompt)

	if prompt == "" {
		bot.sendMessage(chatID, fmt.Sprintf("Укажите задачу после команды. Пример: %s <задача>", c.prefix))
		return
	}

	bot.logger.Info(fmt.Sprintf("[AGENT] agent=%s userID=%d chatID=%d prompt=%s",
		c.agentName, userID, chatID, prompt))
	bot.sendMessage(chatID, fmt.Sprintf("⏳ Запуск агента %s...", c.displayName))

	go func() {
		startTime := time.Now()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(bot.cfg.OpenCodeTimeout)*time.Second)
		defer cancel()

		result, err := bot.agent.ExecuteWithAgent(ctx, prompt, c.agentName)
		elapsed := time.Since(startTime)

		if err != nil {
			bot.logger.Error(fmt.Sprintf("[AGENT] agent=%s userID=%d chatID=%d elapsed=%s error=%v",
				c.agentName, userID, chatID, elapsed, err))
			bot.sendMessage(chatID, fmt.Sprintf("❌ Агент %s завершился с ошибкой: %v", c.displayName, err))
		} else {
			bot.logger.Info(fmt.Sprintf("[AGENT] agent=%s userID=%d chatID=%d elapsed=%s status=success",
				c.agentName, userID, chatID, elapsed))
			if result == "" {
				bot.sendMessage(chatID, fmt.Sprintf("✅ Агент %s выполнил задачу (ответ пуст)", c.displayName))
			} else {
				bot.sendLongMessage(chatID, result)
			}
		}
	}()
}

func registerAgentCommands(router *CommandRouter) {
	agents := []struct {
		prefix      string
		agentName   string
		displayName string
	}{
		// Go-разработка
		{"/go-senior", "go-senior", "Go Senior Developer"},
		{"/go-review", "go-reviewer", "Go Review"},
		{"/debug", "debugging", "Debugging"},
		{"/refactor", "refactoring", "Refactoring"},

		// Базы данных
		{"/postgres", "postgres-sql", "PostgreSQL"},
		{"/clickhouse", "clickhouse-sql", "ClickHouse SQL"},
		{"/dbint", "db-integration", "DB Integration"},

		// Инфраструктура
		{"/bash", "bash-linux", "Bash/Linux"},
		{"/kestra", "kestra", "Kestra"},
		{"/pipeline", "data-pipeline-architect", "Data Pipeline Architect"},
		{"/orchestrate", "orchestrator", "Orchestrator"},

		// Frontend
		{"/react-dev", "react-developer", "React Developer"},
		{"/react-review", "react-reviewer", "React Review"},

		// Документирование и тестирование
		{"/tz", "tz-writer", "ТЗ Writer"},
		{"/docgen", "documentation", "Documentation"},
		{"/write-tests", "test-writer", "Test Writer"},
		{"/explore", "explore", "Explore"},
	}

	for _, a := range agents {
		router.Register(NewAgentCommand(a.prefix, a.agentName, a.displayName))
	}
}
