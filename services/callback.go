package services

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

func GetBonus(botLang string, callbackQuery *tgbotapi.CallbackQuery) error {
	user := auth.GetUser(botLang, callbackQuery.From.ID)

	return user.GetABonus(botLang, callbackQuery)
}

func sendPaymentMethod(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)

	msg := tgbotapi.NewMessage(int64(message.From.ID), assets.LangText(user.Language, "select_payment"))

	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_1"),
			msgs2.NewDataButton("withdrawal_method_2")),
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_3"),
			msgs2.NewDataButton("withdrawal_method_4")),
		msgs2.NewRow(msgs2.NewDataButton("withdrawal_method_5"),
			msgs2.NewDataButton("withdrawal_method_6")),
		msgs2.NewRow(msgs2.NewDataButton("main_back")),
	).Build(user.Language)

	return msgs2.SendMsgToUser(botLang, msg)
}

func CheckSubsAndWithdrawal(botLang string, callBack *tgbotapi.CallbackQuery, userID int) error {
	amount := strings.Split(callBack.Data, "?")[1]

	lang := auth.GetLang(botLang, userID)
	err := msgs2.SendAnswerCallback(botLang, callBack, lang, "invitation_to_subscribe")
	if err != nil {
		return err
	}

	u := auth.GetUser(botLang, userID)
	amountInt, _ := strconv.Atoi(amount)

	member, err := u.CheckSubscribeToWithdrawal(botLang, callBack, userID, amountInt)
	if err != nil {
		return err
	}
	if member {
		db.RdbSetUser(botLang, userID, "main")

		return SendMenu(botLang, userID, assets.LangText(lang, "main_select_menu"))
	}

	return nil
}

func ChangeLanguage(botLang string, callbackQuery *tgbotapi.CallbackQuery) error {
	if db.GetLevel(botLang, callbackQuery.From.ID) != "main" {
		lang := auth.GetLang(botLang, callbackQuery.From.ID)
		return msgs2.SendAnswerCallback(botLang, callbackQuery, lang, "unfinished_action")
	}

	data := strings.Split(callbackQuery.Data, "/")
	if len(data) == 2 {
		err := setLanguage(botLang, callbackQuery.From.ID, data[1])
		if err != nil {
			return err
		}

		err = msgs2.SendAnswerCallback(botLang, callbackQuery, data[1], "language_successful_set")
		if err != nil {
			return err
		}
		deleteTemporaryMessages(botLang, callbackQuery.From.ID)
		return nil
	}

	return sendLanguages(botLang, callbackQuery)
}

func setLanguage(botLang string, userID int, lang string) error {
	db.RdbSetUser(botLang, userID, "main")

	if lang == "back" {
		return SendMenu(botLang, userID, assets.LangText(auth.GetLang(botLang, userID), "back_to_main_menu"))
	}

	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET lang = ? WHERE id = ?;", lang, userID)
	if err != nil {
		text := "Fatal Err with DB - callback.95 //" + err.Error()
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	return SendMenu(botLang, userID, assets.LangText(lang, "back_to_main_menu"))
}

func sendLanguages(botLang string, callbackQuery *tgbotapi.CallbackQuery) error {
	userID := callbackQuery.From.ID
	lang := auth.GetLang(botLang, userID)
	msg := tgbotapi.NewMessage(int64(userID), assets.LangText(lang, "select_language"))

	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("lang_de", "change_lang/de")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("lang_en", "change_lang/en")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("lang_es", "change_lang/es")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("lang_it", "change_lang/it")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("lang_pt", "change_lang/pt")),
		msgs2.NewIlRow(msgs2.NewIlDataButton("back_to_main_menu_button", "change_lang/back")),
	).Build(lang)

	bot := assets.GetBot(botLang)
	data, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}

	err = msgs2.SendAnswerCallback(botLang, callbackQuery, lang, "make_a_choice")
	if err != nil {
		return err
	}

	db.RdbSetTemporary(botLang, userID, data.MessageID)
	return nil
}

func deleteTemporaryMessages(botLang string, userID int) {
	result := db.RdbGetTemporary(botLang, userID)

	if result == "" {
		return
	}

	msgID, err := strconv.Atoi(result)
	if err != nil {
		log.Println(err)
	}

	msg := tgbotapi.NewDeleteMessage(int64(userID), msgID)

	bot := assets.GetBot(botLang)
	if _, err = bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
		log.Println(err)
	}
}
