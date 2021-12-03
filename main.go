package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/log"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services"
	"github.com/Stepan1328/voice-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	rand.Seed(time.Now().Unix())

	logger := log.NewDefaultLogger().Prefix("Voice Bot")
	log.PrintLogo("Voice Bot", []string{"3C91FF"})

	startServices(logger)
	startAllBot(logger)
	assets.UploadUpdateStatistic()

	startHandlers(logger)
}

func startAllBot(log log.Logger) {
	k := 0
	for lang, globalBot := range model.Bots {
		startBot(globalBot, log, lang, k)
		model.Bots[lang].MessageHandler = NewMessagesHandler()
		model.Bots[lang].CallbackHandler = NewCallbackHandler()
		model.Bots[lang].AdminMessageHandler = NewAdminMessagesHandler()
		model.Bots[lang].AdminCallBackHandler = NewAdminCallbackHandler()
		k++
	}

	log.Ok("All bots is running")
}

func startServices(log log.Logger) {
	model.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseAdminMap()
	assets.ParseSiriTasks()
	assets.UploadAdminSettings()
	assets.ParseCommandsList()

	log.Ok("All services are running successfully")
}

func startBot(b *model.GlobalBot, log log.Logger, lang string, k int) {
	var err error
	b.Bot, err = tgbotapi.NewBotAPI(b.BotToken)
	if err != nil {
		log.Fatal("error start bot: %s", err.Error())
	}

	u := tgbotapi.NewUpdate(0)

	b.Chanel = b.Bot.GetUpdatesChan(u)

	b.Rdb = model.StartRedis(k)
	b.DataBase = model.UploadDataBase(lang)
}

func startHandlers(logger log.Logger) {
	wg := new(sync.WaitGroup)

	for botLang, handler := range model.Bots {
		wg.Add(1)
		go func(botLang string, handler *model.GlobalBot, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel, logger)
		}(botLang, handler, wg)
	}

	logger.Ok("All handlers are running")
	_ = msgs.NewParseMessage("it", 872383555, "All bots are restart")
	wg.Wait()
}

func NewMessagesHandler() *services.MessagesHandlers {
	handle := services.MessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewCallbackHandler() *services.CallBackHandlers {
	handle := services.CallBackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminMessagesHandler() *administrator.AdminMessagesHandlers {
	handle := administrator.AdminMessagesHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}

func NewAdminCallbackHandler() *administrator.AdminCallbackHandlers {
	handle := administrator.AdminCallbackHandlers{
		Handlers: map[string]model.Handler{},
	}

	handle.Init()
	return &handle
}
