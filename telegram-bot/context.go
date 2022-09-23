package telegrambot

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Ctx struct {
	values     map[string]any
	ctx        context.Context
	messageID  int
	channelID  int64
	senderID   int64
	senderName string
}

func NewContext(ctx context.Context, message *tgbotapi.Message) *Ctx {
	c := &Ctx{
		values:     make(map[string]any),
		ctx:        ctx,
		messageID:  message.MessageID,
		channelID:  message.Chat.ID,
		senderID:   message.From.ID,
		senderName: message.From.String(),
	}
	return c
}

func (e *Ctx) Locals(key string, val ...any) any {
	if len(val) > 0 {
		e.values[key] = val[0]
		return val[0]
	}
	return e.values[key]
}

func (e *Ctx) SetContext(ctx context.Context) {
	e.ctx = ctx
}

func (e *Ctx) GetContext() context.Context {
	return e.ctx
}
