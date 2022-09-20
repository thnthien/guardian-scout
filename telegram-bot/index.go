package telegrambot

import (
	"log"

	"github.com/thnthien/great-deku/l"
	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Config struct {
	ApiKey                    string  `json:"api_key,omitempty"`
	TimeOut                   int     `json:"time_out,omitempty"`
	Debug                     bool    `json:"debug,omitempty"`
	RestrictMention           bool    `json:"restrict_mention"`
	AllowProcessNormalMessage bool    `json:"allow_process_normal_message"`
	OnlyAllowWhiteList        bool    `json:"only_allow_white_list"`
	UsingBlackList            bool    `json:"using_black_list"`
	WhiteListChatIDs          []int64 `json:"white_list_chat_ids"`
	BlackListChatIDs          []int64 `json:"black_list_chat_ids"`
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
	ll  l.Logger
	cfg Config

	whiteListMap map[int64]any
	blackListMap map[int64]any

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

	return teleBot
}

func (b *TeleBot) SetLogger(logger *zap.Logger) {
	b.ll = l.Logger{Logger: logger}
}
