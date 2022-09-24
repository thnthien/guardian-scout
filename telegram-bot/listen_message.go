package telegrambot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thnthien/great-deku/l"
)

func (b *TeleBot) ListenMessage() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = b.cfg.TimeOut
	if u.Timeout == 0 {
		u.Timeout = 60
	}

	bot, err := b.GetMe()
	if err != nil {
		b.ll.Fatal("cannot get bot info", l.Error(err))
	}
	b.ll.Info("started bot", l.Object("bot_info", bot))
	updates := b.GetUpdatesChan(u)

	ctx := context.Background()
	for update := range updates {
		if update.Message == nil {
			continue
		}

		message := update.Message

		if b.cfg.UsingBlackList {
			if _, ok := b.blackListMap[message.Chat.ID]; ok {
				continue
			}
		}

		if b.cfg.OnlyAllowWhiteList {
			if _, ok := b.whiteListMap[message.Chat.ID]; !ok {
				continue
			}
		}

		if !message.IsCommand() && !b.cfg.AllowProcessNormalMessage {
			continue
		}

		command := ""
		if message.IsCommand() {
			command = message.Command()

			if strings.Contains(message.Text, fmt.Sprintf("/%s@", command)) && b.cfg.RestrictMention {
				at := strings.Split(message.CommandWithAt(), "@")[1]
				if at != "" && at != bot.UserName {
					b.ll.Debug("command for other bots, ignore")
					continue
				}
			}
		}

		handlers, ok := b.handlers[command]
		if !ok {
			if !b.cfg.AllowProcessNormalMessage {
				continue
			}
			handlers = b.defaultHandler
		}

		b.rpooling.Submit(b.processMessage(newContext(ctx, b.ll, b, &bot, message, handlers)))
	}

	return nil
}

func (b *TeleBot) processMessage(c *Ctx) func() {
	return func() {
		if err := c.Next(); err != nil && b.errorHandler != nil {
			b.errorHandler(c, err)
		}
	}
}
