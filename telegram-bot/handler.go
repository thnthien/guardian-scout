package telegrambot

import (
	"errors"
)

type Handler func(ctx *Ctx) error

func (b *TeleBot) RegisterHandler(command string, handlers ...Handler) error {
	if b.handlers == nil {
		b.handlers = make(map[string][]Handler)
	}

	if command == "" {
		return errors.New("command cannot be an empty string")
	}

	b.handlers[command] = make([]Handler, 0, len(handlers))
	for i := len(handlers) - 1; i >= 0; i-- {
		b.handlers[command] = append(b.handlers[command], handlers[i])
	}
	return nil
}

func (b *TeleBot) SetDefaultHandler(handlers ...Handler) {
	b.defaultHandler = handlers
}
