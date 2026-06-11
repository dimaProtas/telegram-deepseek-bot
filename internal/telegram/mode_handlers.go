package telegram

import (
	"fmt"
	"strings"

	"telegram-deepseek-bot/internal/deepseek"
)

func (b *Bot) handleMode(chatID int64, text string) {
	parts := strings.Split(text, " ")
	if len(parts) < 2 {
		b.sendMessage(chatID, "Пожалуйста, укажите модель: chat, coder или reasoner.\nПример: /mode coder")
		return
	}

	model := strings.ToLower(parts[1])
	switch model {
	case "chat", "coder", "reasoner":
		b.store.SetModel(chatID, model)
		b.sendMessage(chatID, fmt.Sprintf("Модель изменена на: %s", model))
	default:
		b.sendMessage(chatID, "Неизвестная модель. Доступные: chat, coder, reasoner.")
	}
}

func (b *Bot) handleCode(chatID int64, prompt string) {
	// Для /code принудительно используем coder модель, но сохраняем контекст
	b.store.AddMessage(chatID, "user", prompt)

	history := b.store.GetMessages(chatID)
	var messages []deepseek.Message
	for _, m := range history {
		messages = append(messages, deepseek.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	req := deepseek.ChatRequest{
		Model:     "deepseek-coder",
		Messages:  messages,
		Temp:      b.cfg.DeepSeekTemp,
		MaxTokens: b.cfg.DeepSeekMaxTokens,
		Stream:    false,
	}

	b.logger.Info(fmt.Sprintf("Sending code request for chatID %d", chatID))
	resp, err := b.ds.Chat(req)
	if err != nil {
		b.logger.Error(fmt.Sprintf("DeepSeek API error: %v", err))
		b.sendMessage(chatID, "Произошла ошибка при генерации кода.")
		return
	}

	if len(resp.Choices) == 0 {
		b.sendMessage(chatID, "AI не вернул код.")
		return
	}

	answer := resp.Choices[0].Message.Content
	b.store.AddMessage(chatID, "assistant", answer)
	b.sendMessage(chatID, answer)
}
