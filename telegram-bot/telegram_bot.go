package telegrambot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/thnthien/great-deku/l"
	"github.com/thnthien/great-deku/rpooling"
	"go.uber.org/zap"
)

type Config struct {
	//ApiKey secret api key of telegram bot
	ApiKey string `json:"api_key,omitempty"`

	//TimeOut connection timeout of bot (seconds)
	TimeOut int `json:"time_out,omitempty"`

	Debug bool `json:"debug,omitempty"`

	//RestrictMention if this true, and the command has mention inside, only the right mention of bot can be processed:
	//eg: /command@the_bot_user_name
	RestrictMention bool `json:"restrict_mention"`

	//AllowProcessNormalMessage if this true, so the bot will process all message. By normal, only command message can be passed to bot
	AllowProcessNormalMessage bool `json:"allow_process_normal_message"`

	//OnlyAllowWhiteList using the whitelist chat ids to filter
	OnlyAllowWhiteList bool `json:"only_allow_white_list"`

	//UsingBlackList using the black list chat ids to filter
	UsingBlackList bool `json:"using_black_list"`

	//MaxThreadNumber how many goroutine can be run at the same time
	MaxThreadNumber int `json:"max_thread_number"`

	//WhiteListChatIDs whitelist chat ids
	WhiteListChatIDs []int64 `json:"white_list_chat_ids"`

	//BlackListChatIDs blacklist chat ids
	BlackListChatIDs []int64 `json:"black_list_chat_ids"`

	//ErrorHandler
	ErrorHandler ErrorHandler
}

var DefaultConfig = Config{
	TimeOut:                   60,
	Debug:                     false,
	RestrictMention:           true,
	AllowProcessNormalMessage: false,
	OnlyAllowWhiteList:        false,
	UsingBlackList:            false,
	WhiteListChatIDs:          nil,
	BlackListChatIDs:          nil,
}

type TeleBot struct {
	*tgbotapi.BotAPI

	ll       l.Logger
	cfg      Config
	rpooling rpooling.IPool

	whiteListMap map[int64]any
	blackListMap map[int64]any

	middlewares    []Handler
	handlers       map[string][]Handler
	defaultHandler []Handler
	errorHandler   ErrorHandler
}

// region init
func New(apiKey string, cfgs ...Config) *TeleBot {
	cfg := DefaultConfig
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	cfg.ApiKey = apiKey

	bot, err := tgbotapi.NewBotAPI(cfg.ApiKey)
	if err != nil {
		log.Fatalf("cannot create bot: %v", err)
	}

	return initBot(bot, l.New(), cfg)
}

func NewWithTelegramBot(bot *tgbotapi.BotAPI, cfgs ...Config) *TeleBot {
	cfg := DefaultConfig
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return initBot(bot, l.New(), cfg)
}

func initBot(bot *tgbotapi.BotAPI, ll l.Logger, cfg Config) *TeleBot {
	teleBot := &TeleBot{
		BotAPI: bot,
		ll:     ll,
		cfg:    cfg,
	}
	bot.Debug = cfg.Debug

	if cfg.OnlyAllowWhiteList {
		teleBot.whiteListMap = make(map[int64]any)
		for _, chatID := range cfg.WhiteListChatIDs {
			teleBot.whiteListMap[chatID] = nil
		}
	}
	if cfg.UsingBlackList {
		teleBot.blackListMap = make(map[int64]any)
		for _, chatID := range cfg.BlackListChatIDs {
			teleBot.blackListMap[chatID] = nil
		}
	}
	if cfg.MaxThreadNumber == 0 {
		cfg.MaxThreadNumber = 1000
	}
	teleBot.rpooling = rpooling.New(cfg.MaxThreadNumber, ll)

	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = NewDefaultErrorHandler()
	}
	teleBot.errorHandler = cfg.ErrorHandler

	return teleBot
}

// SetLogger set logger for bot, by default, bot always init it's own logger
func (b *TeleBot) SetLogger(logger *zap.Logger) {
	b.ll = l.Logger{Logger: logger}
}

//endregion

// Use middleware adding to the root
func (b *TeleBot) Use(h Handler) {
	if b.middlewares == nil {
		b.middlewares = make([]Handler, 0)
	}
	b.middlewares = append([]Handler{h}, b.middlewares...)
}

// SendTextMessage send response message
func (b *TeleBot) SendTextMessage(chatID int64, text string, replyMessageID ...int) ([]*tgbotapi.Message, error) {
	sentMsgs := make([]*tgbotapi.Message, 0)
	for i := 0; i < len(text); i += 4096 {
		length := i + 4096
		if length > len(text) {
			length = len(text)
		}
		message := tgbotapi.NewMessage(chatID, text[i:length])
		message.ParseMode = "Markdown"
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

// ListenMessage start listen message from telegram
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

		if len(b.middlewares) > 0 {
			handlers = append(b.middlewares, handlers...)
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

// region handlers
type Handler func(ctx *Ctx) error

// RegisterHandler add handlers for handle command
// eg: for the command /foo, we can register like b.RegisterHandler("foo", handler1, handler2)
// Note: the order of executing is from right to left --> handler2 first than handler1
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

// SetDefaultHandler default handlers for case of process normal message
func (b *TeleBot) SetDefaultHandler(handlers ...Handler) {
	b.defaultHandler = handlers
}

//endregion
