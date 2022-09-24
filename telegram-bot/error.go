package telegrambot

import (
	"fmt"

	"github.com/thnthien/great-deku/l"
)

type ErrorHandler func(c *Ctx, err error)

func NewDefaultErrorHandler() ErrorHandler {
	messageTemplate := "error when process message: ```%s```"
	return func(c *Ctx, err error) {
		if _, e := c.bot.SendTextMessage(c.channelID, fmt.Sprintf(messageTemplate, err), c.messageID); e != nil {
			c.ll.Error("cannot send error message", l.Error(e))
		}
	}
}
