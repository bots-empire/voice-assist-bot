package assets

import (
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	Bots = make(map[string]Handler)
)

type Handler struct {
	Chanel tgbotapi.UpdatesChannel
	Bot    *tgbotapi.BotAPI
	Rdb    *redis.Client
}

func GetBot(botLang string) *tgbotapi.BotAPI {
	return Bots[botLang].Bot
}
