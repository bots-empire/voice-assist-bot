package main

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
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
	assets.UploadUpdateStatistic()

	startHandlers()
}

func startAllBot() {
	k := 0
	for lang, bot := range cfg.Tokens {
		assets.Bots[lang] = startBot(bot.Token, lang, k)
		k++
	}

	log.Println("All bots is running")
}

func startBot(botToken, lang string, k int) assets.Handler {
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
	dataBase := db.UploadDataBase(lang)
	dataBase.SetMaxOpenConns(100)

	return assets.Handler{
		Chanel:   updates,
		Bot:      bot,
		Rdb:      rdb,
		DataBase: dataBase,
	}
}

func startServices() {
	cfg.FillBotsConfig()
	assets.ParseLangMap()
	assets.ParseSiriTasks()
	assets.ParseAdminMap()
	assets.UploadAdminSettings()

	log.Println("All services are running successfully")
}

func startHandlers() {
	wg := &sync.WaitGroup{}

	for botLang, handler := range assets.Bots {
		wg.Add(1)
		go func(botLang string, handler assets.Handler, wg *sync.WaitGroup) {
			defer wg.Done()

			//testChanel := make(chan tgbotapi.Update)
			//go func() {
			//	time.Sleep(5 * time.Second)
			//
			//	dur := 5 * time.Millisecond
			//	for {
			//		testUpdate := tgbotapi.Update{
			//			Message: &tgbotapi.Message{
			//				From: &tgbotapi.User{
			//					ID:           1418862576,
			//					FirstName:    "Stepan",
			//					LastName:     "Samsonov",
			//					UserName:     "ridpiner",
			//					LanguageCode: "ru",
			//				},
			//				Chat: &tgbotapi.Chat{
			//					ID: 1418862576,
			//				},
			//				Text: "/start",
			//			},
			//		}
			//		testChanel <- testUpdate
			//		time.Sleep(dur)
			//	}
			//}()
			services.ActionsWithUpdates(botLang, handler.Chanel)
		}(botLang, handler, wg)
	}

	log.Println("All handlers are running")
	_ = msgs2.NewParseMessage("it", 1418862576, "All bots are restart")
	wg.Wait()
}
