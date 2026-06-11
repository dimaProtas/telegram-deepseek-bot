package telegram

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Command interface {
	Name() string
	Matches(text string) bool
	Execute(bot *Bot, msg *tgbotapi.Message)
}

type CommandRouter struct {
	commands []Command
}

func NewCommandRouter() *CommandRouter {
	return &CommandRouter{}
}

func (r *CommandRouter) Register(cmd Command) {
	r.commands = append(r.commands, cmd)
}

func (r *CommandRouter) Dispatch(bot *Bot, msg *tgbotapi.Message) bool {
	text := msg.Text
	for _, cmd := range r.commands {
		if cmd.Matches(text) {
			cmd.Execute(bot, msg)
			return true
		}
	}
	return false
}

type exactCommand struct {
	name   string
	prefix string
	fn     func(bot *Bot, msg *tgbotapi.Message)
}

func (c *exactCommand) Name() string                          { return c.name }
func (c *exactCommand) Matches(text string) bool              { return strings.HasPrefix(text, c.prefix) }
func (c *exactCommand) Execute(bot *Bot, msg *tgbotapi.Message) { c.fn(bot, msg) }

type prefixCommand struct {
	name   string
	prefix string
	fn     func(bot *Bot, msg *tgbotapi.Message, arg string)
}

func (c *prefixCommand) Name() string             { return c.name }
func (c *prefixCommand) Matches(text string) bool { return strings.HasPrefix(text, c.prefix+" ") || strings.HasPrefix(text, c.prefix+"@") }
func (c *prefixCommand) Execute(bot *Bot, msg *tgbotapi.Message) {
	arg := strings.TrimPrefix(msg.Text, c.prefix)
	arg = strings.TrimPrefix(arg, "@"+bot.api.Self.UserName)
	arg = strings.TrimSpace(arg)
	c.fn(bot, msg, arg)
}

type fallbackCommand struct {
	name string
	fn   func(bot *Bot, msg *tgbotapi.Message)
}

func (c *fallbackCommand) Name() string                          { return c.name }
func (c *fallbackCommand) Matches(text string) bool              { return true }
func (c *fallbackCommand) Execute(bot *Bot, msg *tgbotapi.Message) { c.fn(bot, msg) }
