package services

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

func GetBonus(callbackQuery *tgbotapi.CallbackQuery) {
	user := auth.GetUser(callbackQuery.From.ID)

	user.GetABonus()
}

func Withdrawal(callbackQuery *tgbotapi.CallbackQuery) {
	level := auth.GetLevel(callbackQuery.From.ID)
	if level != "main" {
		lang := auth.GetLang(callbackQuery.From.ID)
		notice := tgbotapi.CallbackConfig{
			CallbackQueryID: callbackQuery.ID,
			Text:            assets.LangText(lang, "unfinished_action"),
		}

		if _, err := assets.Bot.AnswerCallbackQuery(notice); err != nil {
			log.Println(err)
		}
		return
	}

	stringID := strconv.Itoa(callbackQuery.From.ID)
	_, err := db.Rdb.Set(stringID, "withdrawal", 0).Result()
	if err != nil {
		log.Println(err)
	}

	sendPaymentMethod(callbackQuery)
}

func sendPaymentMethod(callbackQuery *tgbotapi.CallbackQuery) {
	user := auth.GetUser(callbackQuery.From.ID)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, assets.LangText(user.Language, "select_payment"))

	payPal := tgbotapi.NewKeyboardButton(assets.LangText(user.Language, "paypal_method"))
	creditCard := tgbotapi.NewKeyboardButton(assets.LangText(user.Language, "credit_card_method"))
	row1 := tgbotapi.NewKeyboardButtonRow(payPal, creditCard)

	back := tgbotapi.NewKeyboardButton(assets.LangText(user.Language, "main_back"))
	row2 := tgbotapi.NewKeyboardButtonRow(back)

	markUp := tgbotapi.NewReplyKeyboard(row1, row2)
	msg.ReplyMarkup = markUp

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

//func sendSimpleMsg(chatID int64, text string) {
//	msg := tgbotapi.NewMessage(chatID, text)
//
//	if _, err := assets.Bot.Send(msg); err != nil {
//		log.Println(err)
//	}
//}
