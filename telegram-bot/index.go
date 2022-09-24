package telegrambot

import (
	"log"

	"github.com/thnthien/great-deku/l"
	"github.com/thnthien/great-deku/rpooling"
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
	MaxThreadNumber           int     `json:"max_thread_number"`
	WhiteListChatIDs          []int64 `json:"white_list_chat_ids"`
	BlackListChatIDs          []int64 `json:"black_list_chat_ids"`
	ErrorHandler              ErrorHandler
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

	handlers       map[string][]Handler
	defaultHandler []Handler
	errorHandler   ErrorHandler
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

func (b *TeleBot) SetLogger(logger *zap.Logger) {
	b.ll = l.Logger{Logger: logger}
}
