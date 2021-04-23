package services

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	"github.com/Stepan1328/voice-assist-bot/services/msgs"
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
	level := db.GetLevel(callbackQuery.From.ID)
	if level != "main" && level != "empty" {
		lang := auth.GetLang(callbackQuery.From.ID)
		msgs.SendAnswerCallback(callbackQuery, lang, "unfinished_action")
		return
	}

	db.RdbSetUser(callbackQuery.From.ID, "withdrawal")

	sendPaymentMethod(callbackQuery)
}

func sendPaymentMethod(callbackQuery *tgbotapi.CallbackQuery) {
	user := auth.GetUser(callbackQuery.From.ID)

	msg := tgbotapi.NewMessage(callbackQuery.Message.Chat.ID, assets.LangText(user.Language, "select_payment"))

	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow("paypal_method", "credit_card_method"),
		msgs.NewRow("main_back"),
	).Build(user.Language)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func ChangeLanguage(callbackQuery *tgbotapi.CallbackQuery) {
	if db.GetLevel(callbackQuery.From.ID) != "main" {
		lang := auth.GetLang(callbackQuery.From.ID)
		msgs.SendAnswerCallback(callbackQuery, lang, "unfinished_action")
		return
	}

	data := strings.Split(callbackQuery.Data, "/")
	if len(data) == 2 {
		setLanguage(callbackQuery.From.ID, data[1])
		msgs.SendAnswerCallback(callbackQuery, data[1], "language_successful_set")
		deleteTemporaryMessages(callbackQuery.From.ID)
		return
	}

	sendLanguages(callbackQuery)
}

func setLanguage(userID int, lang string) {
	db.RdbSetUser(userID, "main")

	if lang == "back" {
		SendMenu(userID, assets.LangText(auth.GetLang(userID), "back_to_main_menu"))
		return
	}
	_, err := db.DataBase.Query("UPDATE users SET lang = ? WHERE id = ?;", lang, userID)
	if err != nil {
		panic(err.Error())
	}

	SendMenu(userID, assets.LangText(lang, "back_to_main_menu"))
}

func sendLanguages(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := auth.GetLang(userID)
	msg := tgbotapi.NewMessage(int64(userID), assets.LangText(lang, "select_language"))

	msg.ReplyMarkup = msgs.NewInlineMarkUp(
		msgs.NewInlineRow(msgs.NewDataButton("lang_de", "change_lang/de")),
		msgs.NewInlineRow(msgs.NewDataButton("lang_en", "change_lang/en")),
		msgs.NewInlineRow(msgs.NewDataButton("lang_es", "change_lang/es")),
		msgs.NewInlineRow(msgs.NewDataButton("lang_it", "change_lang/it")),
		msgs.NewInlineRow(msgs.NewDataButton("lang_pt", "change_lang/pt")),
		msgs.NewInlineRow(msgs.NewDataButton("back_to_main_menu_button", "change_lang/back")),
	).Build(lang)

	data, err := assets.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	msgs.SendAnswerCallback(callbackQuery, lang, "make_a_choice")

	db.RdbSetTemporary(userID, data.MessageID)
}

func deleteTemporaryMessages(userID int) {
	result := db.RdbGetTemporary(userID)

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
