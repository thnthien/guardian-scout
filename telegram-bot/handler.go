package telegrambot

import (
	"errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler func(ctx *Ctx, message *tgbotapi.Message)

func (b *TeleBot) RegisterHandler(command string, handler Handler) error {
	if b.handlers == nil {
		b.handlers = make(map[string]Handler)
	}

	if command == "" {
		return errors.New("command cannot be an empty string")
	}

	b.handlers[command] = handler
	return nil
}

func (b *TeleBot) SetHandlerMap(handlers map[string]Handler) error {
	for key, _ := range handlers {
		if key == "" {
			return errors.New("command cannot be an empty string")
		}
	}
	b.handlers = handlers
	return nil
}

func (b *TeleBot) SetDefaultHandler(handler Handler) {
	b.defaultHandler = handler
}
