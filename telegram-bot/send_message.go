package telegrambot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thnthien/great-deku/l"
)

func (b *TeleBot) SendTextMessage(chatID int64, text string, replyMessageID ...int) ([]*tgbotapi.Message, error) {
	sentMsgs := make([]*tgbotapi.Message, 0)
	for i := 0; i < len(text); i += 4096 {
		length := i + 4096
		if length > len(text) {
			length = len(text)
		}
		message := tgbotapi.NewMessage(chatID, text[i:length])
		message.ParseMode = "HTML"
		if len(replyMessageID) > 0 {
			message.ReplyToMessageID = replyMessageID[0]
		}
		sentMsg, err := b.Send(message)
		if err != nil {
			b.ll.Error("cannot sent message", l.Object("message", message), l.Error(err))
			return nil, err
		}
		sentMsgs = append(sentMsgs, &sentMsg)
	}
	return sentMsgs, nil
}
