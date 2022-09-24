package telegrambot

import (
	"context"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/satori/go.uuid"
	"github.com/thnthien/great-deku/l"
)

var spaceRe = regexp.MustCompile("\\s+")

type Ctx struct {
	ctx       context.Context
	ll        l.Logger
	bot       *TeleBot
	botInfo   *tgbotapi.User
	values    map[string]any
	messageID int
	channelID int64
	requestID string
	sender    *tgbotapi.User
	isCommand bool
	message   *tgbotapi.Message
	params    []string

	handlers     []Handler
	handlerIndex int
}

func newContext(
	ctx context.Context, ll l.Logger, bot *TeleBot, botInfo *tgbotapi.User,
	msg *tgbotapi.Message, handlers []Handler) *Ctx {
	c := &Ctx{
		ctx:          ctx,
		ll:           ll,
		bot:          bot,
		botInfo:      botInfo,
		values:       make(map[string]any),
		messageID:    msg.MessageID,
		channelID:    msg.Chat.ID,
		requestID:    uuid.NewV4().String(),
		sender:       msg.From,
		isCommand:    msg.IsCommand(),
		message:      msg,
		handlers:     handlers,
		handlerIndex: 0,
	}

	initCtx(c, msg)

	return c
}

func initCtx(c *Ctx, msg *tgbotapi.Message) {
	if msg.CommandArguments() != "" {
		c.params = spaceRe.Split(msg.CommandArguments(), -1)
	}
}

func (c *Ctx) Locals(key string, val ...any) any {
	if len(val) > 0 {
		c.values[key] = val[0]
		return val[0]
	}
	return c.values[key]
}

func (c *Ctx) SetContext(ctx context.Context) {
	c.ctx = ctx
}

func (c *Ctx) GetContext() context.Context {
	return c.ctx
}

func (c *Ctx) GetSender() *tgbotapi.User {
	return c.sender
}

func (c *Ctx) SenderID() int64 {
	if c.sender == nil {
		return 0
	}
	return c.sender.ID
}

func (c *Ctx) SenderName() string {
	if c.sender == nil {
		return ""
	}
	return c.sender.String()
}

func (c *Ctx) Message() *tgbotapi.Message {
	return c.message
}

func (c *Ctx) ChannelID() int64 {
	return c.channelID
}

func (c *Ctx) MessageID() int {
	return c.messageID
}

func (c *Ctx) IsCommand() bool {
	return c.isCommand
}

func (c *Ctx) GetRequestID() string {
	return c.requestID
}

func (c *Ctx) Next() error {
	handler := c.handlers[c.handlerIndex]
	c.handlerIndex++
	return handler(c)
}
