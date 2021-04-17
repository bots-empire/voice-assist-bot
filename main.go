package main

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os"
)

func main() {
	var updates tgbotapi.UpdatesChannel

	db.UploadDataBase()
	assets.ParseLangMap()
	services.Bot, updates = startBot()

	services.ActionsWithUpdates(updates)
}

func startBot() (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel) {
	botToken := takeBotToken()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		panic("Failed to initialize bot: " + err.Error())
	}

	log.Println("The bot is running")

	return bot, updates
}

func takeBotToken() string {
	content, _ := os.ReadFile("./botToken.txt")
	return string(content)
}
