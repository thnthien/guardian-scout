package telegrambot

import (
	"log"

	"github.com/thnthien/great-deku/l"
	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	ApiKey                    string `json:"api_key,omitempty"`
	TimeOut                   int    `json:"time_out,omitempty"`
	Debug                     bool   `json:"debug,omitempty"`
	RestrictMention           bool   `json:"restrict_mention"`
	AllowProcessNormalMessage bool   `json:"allow_process_normal_message"`
}

var DefaultConfig = Config{
	TimeOut:                   60,
	Debug:                     false,
	RestrictMention:           true,
	AllowProcessNormalMessage: false,
}

type TeleBot struct {
	*tgbotapi.BotAPI
	ll  l.Logger
	cfg Config

	handlers       map[string]Handler
	defaultHandler Handler
}

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
	bot.Debug = cfg.Debug
	teleBot := &TeleBot{
		BotAPI: bot,
		ll:     l.New(),
		cfg:    cfg,
	}
	return teleBot
}

func NewWithTelegramBot(bot *tgbotapi.BotAPI, cfgs ...Config) *TeleBot {
	cfg := DefaultConfig
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	return &TeleBot{
		BotAPI: bot,
		ll:     l.New(),
		cfg:    cfg,
	}
}

func (b *TeleBot) SetLogger(logger *zap.Logger) {
	b.ll = l.Logger{Logger: logger}
}
