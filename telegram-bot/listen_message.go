package telegrambot

import (
	"context"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thnthien/great-deku/l"
)

func (b *TeleBot) ListenMessage(ctx context.Context) error {
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

		b.ll.Debug("received message", l.Object("message", message))

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

		handler, ok := b.handlers[command]
		if ok {
			handler(message)
			continue
		}

		if b.cfg.AllowProcessNormalMessage {
			b.defaultHandler(message)
		}
	}

	return nil
}
