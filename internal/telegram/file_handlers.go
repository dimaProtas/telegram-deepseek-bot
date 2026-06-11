package telegram

import (
	"fmt"

	"telegram-deepseek-bot/internal/deepseek"
)

func (b *Bot) handleFileCommand(chatID int64, fileName string, action string) {
	content, err := b.agent.ReadFile(fileName)
	if err != nil {
		b.sendMessage(chatID, fmt.Sprintf("Ошибка чтения файла: %v", err))
		return
	}

	var prompt string
	switch action {
	case "review":
		prompt = fmt.Sprintf("Please perform a professional code review of the following file:\n\n%s", content)
	case "explain":
		prompt = fmt.Sprintf("Please explain what the following code does in detail:\n\n%s", content)
	case "test":
		prompt = fmt.Sprintf("Please generate comprehensive unit tests for the following code:\n\n%s", content)
	case "docs":
		prompt = fmt.Sprintf("Please generate professional documentation for the following code:\n\n%s", content)
	default:
		b.sendMessage(chatID, "Неизвестное действие.")
		return
	}

	b.logger.Info(fmt.Sprintf("Executing %s for file %s (chatID %d)", action, fileName, chatID))
	
	req := deepseek.ChatRequest{
		Model:     "deepseek-coder",
		Messages: []deepseek.Message{
			{Role: "user", Content: prompt},
		},
		Temp:      b.cfg.DeepSeekTemp,
		MaxTokens: b.cfg.DeepSeekMaxTokens,
		Stream:    false,
	}

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

	b.sendMessage(chatID, resp.Choices[0].Message.Content)
}
