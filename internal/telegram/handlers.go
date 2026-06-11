package telegram

import (
	"fmt"

	"telegram-deepseek-bot/internal/deepseek"
)

func (b *Bot) handleChat(chatID int64, prompt string) {
	// Сохраняем сообщение пользователя
	b.store.AddMessage(chatID, "user", prompt)

	// Получаем модель пользователя или дефолтную
	model := b.store.GetModel(chatID, b.cfg.DeepSeekModel)

	// Получаем историю для контекста
	history := b.store.GetMessages(chatID)
	var messages []deepseek.Message
	for _, m := range history {
		messages = append(messages, deepseek.Message{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Запрос к API DeepSeek
	req := deepseek.ChatRequest{
		Model:     model,
		Messages:  messages,
		Temp:      b.cfg.DeepSeekTemp,
		MaxTokens: b.cfg.DeepSeekMaxTokens,
		Stream:    false,
	}

	b.logger.Info(fmt.Sprintf("Sending request to DeepSeek (model: %s) for chatID %d", model, chatID))
	resp, err := b.ds.Chat(req)
	if err != nil {
		b.logger.Error(fmt.Sprintf("DeepSeek API error: %v", err))
		b.sendMessage(chatID, "Произошла ошибка при обращении к API.")
		return
	}

	if len(resp.Choices) == 0 {
		b.sendMessage(chatID, "AI не вернул ответ.")
		return
	}

	// Учитываем токены
	b.store.AddTokens(chatID, resp.Usage.TotalTokens)

	answer := resp.Choices[0].Message.Content
	b.logger.Info(fmt.Sprintf("Received answer for chatID %d", chatID))
	
	// Сохраняем ответ AI
	b.store.AddMessage(chatID, "assistant", answer)
	
	b.sendMessage(chatID, answer)
}
