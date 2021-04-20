package services

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
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

	userID := auth.UserIDToRdb(callbackQuery.From.ID)
	_, err := db.Rdb.Set(userID, "withdrawal", 0).Result()
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

func ChangeLanguage(callbackQuery *tgbotapi.CallbackQuery) {
	if auth.GetLevel(callbackQuery.From.ID) != "main" {
		answerCallback := tgbotapi.CallbackConfig{
			CallbackQueryID: callbackQuery.ID,
			Text:            assets.LangText(auth.GetLang(callbackQuery.From.ID), "unfinished_action"),
		}
		if _, err := assets.Bot.AnswerCallbackQuery(answerCallback); err != nil {
			log.Println(err)
		}
		return
	}

	data := strings.Split(callbackQuery.Data, "/")
	if len(data) == 2 {
		setLanguage(callbackQuery.From.ID, data[1])
		answerCallback := tgbotapi.CallbackConfig{
			CallbackQueryID: callbackQuery.ID,
			Text:            assets.LangText(data[1], "language_successful_set"),
		}
		if _, err := assets.Bot.AnswerCallbackQuery(answerCallback); err != nil {
			log.Println(err)
		}
		deleteTemporaryMessages(callbackQuery.From.ID)
		return
	}

	sendLanguages(callbackQuery)
}

func setLanguage(userID int, lang string) {
	if lang == "back" {
		stringUserID := auth.UserIDToRdb(userID)
		_, err := db.Rdb.Del(stringUserID).Result()
		if err != nil {
			log.Println(err)
		}

		SendMenu(userID, assets.LangText(auth.GetLang(userID), "back_to_main_menu"))
	}

	_, err := db.DataBase.Query("UPDATE users SET lang = ? WHERE id = ?;", lang, userID)
	if err != nil {
		panic(err.Error())
	}

	userIDToRdb := auth.UserIDToRdb(userID)
	_, err = db.Rdb.Del(userIDToRdb).Result()
	if err != nil {
		log.Println(err)
	}
}

func sendLanguages(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := auth.GetLang(userID)
	msg := tgbotapi.NewMessage(int64(userID), assets.LangText(lang, "select_language"))

	en := tgbotapi.NewInlineKeyboardButtonData("English 🇬🇧", "change_lang/en")
	de := tgbotapi.NewInlineKeyboardButtonData("Deutsch 🇩🇪", "change_lang/de")
	es := tgbotapi.NewInlineKeyboardButtonData("Español 🇪🇸", "change_lang/es")
	it := tgbotapi.NewInlineKeyboardButtonData("Italiano 🇮🇹", "change_lang/it")
	pt := tgbotapi.NewInlineKeyboardButtonData("Português 🇵🇹", "change_lang/pt")
	back := tgbotapi.NewInlineKeyboardButtonData(
		assets.LangText(lang, "back_to_main_menu_button"), "change_lang/back")
	msg.ReplyMarkup = createMarkUpFromButton(en, de, es, it, pt, back)

	data, err := assets.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.LangText(lang, "make_a_choice"),
	}
	if _, err = assets.Bot.AnswerCallbackQuery(answerCallback); err != nil {
		log.Println(err)
	}

	temporaryID := auth.TemporaryIDToRdb(userID)
	msgID := data.MessageID
	_, err = db.Rdb.Set(temporaryID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func deleteTemporaryMessages(userID int) {
	temporaryID := auth.TemporaryIDToRdb(userID)
	result, err := db.Rdb.Get(temporaryID).Result()
	if err != nil {
		log.Println(err)
	}

	if result == "" {
		return
	}

	msgID, err := strconv.Atoi(result)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewDeleteMessage(int64(userID), msgID)
	if _, err = assets.Bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
		log.Println(err)
	}
}

func createMarkUpFromButton(buttons ...tgbotapi.InlineKeyboardButton) tgbotapi.InlineKeyboardMarkup {
	var markUp tgbotapi.InlineKeyboardMarkup
	for _, elem := range buttons {
		row := tgbotapi.NewInlineKeyboardRow(elem)
		markUp.InlineKeyboard = append(markUp.InlineKeyboard, row)
	}
	return markUp
}
