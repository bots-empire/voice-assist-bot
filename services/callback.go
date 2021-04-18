package services

import (
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func GetBonus(callbackQuery *tgbotapi.CallbackQuery) {
	user := auth.GetUser(callbackQuery.From.ID)

	user.GetABonus()
}

//func sendSimpleMsg(chatID int64, text string) {
//	msg := tgbotapi.NewMessage(chatID, text)
//
//	if _, err := assets.Bot.Send(msg); err != nil {
//		log.Println(err)
//	}
//}
