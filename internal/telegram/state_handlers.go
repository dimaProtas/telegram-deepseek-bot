package telegram

import (
	"fmt"
)

func (b *Bot) handleState(chatID int64) {
	tokens := b.store.GetTokens(chatID)
	b.sendMessage(chatID, fmt.Sprintf("Ваш общий расход токенов: %d", tokens))
}

func (b *Bot) handleRetry(chatID int64) {
	messages := b.store.GetMessages(chatID)
	if len(messages) < 2 {
		b.sendMessage(chatID, "Нечего повторять.")
		return
	}

	// Последнее сообщение пользователя
	var lastUserMsg string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastUserMsg = messages[i].Content
			break
		}
	}

	if lastUserMsg == "" {
		b.sendMessage(chatID, "Последний запрос не найден.")
		return
	}

	b.logger.Info(fmt.Sprintf("Retrying last request for chatID %d: %s", chatID, lastUserMsg))
	b.handleChat(chatID, lastUserMsg)
}
