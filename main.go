package main

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().Unix())

	startServices()
	startAllBot()

	startHandlers()
}

func startAllBot() {
	k := 0
	for lang, bot := range cfg.Tokens {
		assets.Bots[lang] = startBot(bot.Token, k)
		k++
	}

	log.Println("All bots is running")
}

func startBot(botToken string, k int) assets.Handler {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Println(err)
		return assets.Handler{}
	}

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic("Failed to initialize bot: " + err.Error())
	}

	rdb := db.StartRedis(k)
	return assets.Handler{
		Chanel: updates,
		Bot:    bot,
		Rdb:    rdb,
	}
}

func startServices() {
	cfg.FillBotsConfig()
	db.UploadDataBase()
	assets.ParseLangMap()
	assets.ParseSiriTasks()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()

	log.Println("All services are running successfully")
}

func startHandlers() {
	wg := new(sync.WaitGroup)

	for botLang, handler := range assets.Bots {
		wg.Add(1)
		go func(botLang string, handler assets.Handler, wg *sync.WaitGroup) {
			defer wg.Done()
			services.ActionsWithUpdates(botLang, handler.Chanel)
		}(botLang, handler, wg)
	}

	log.Println("All handlers are running")
	wg.Wait()
}
