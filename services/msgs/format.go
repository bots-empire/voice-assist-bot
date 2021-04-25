package msgs

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func NewParseMessage(chatID int64, text string) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func NewParseMarkUpMessage(chatID int64, markUp interface{}, text string) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func NewIDParseMarkUpMessage(chatID int64, markUp interface{}, text string) int {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	message, err := assets.Bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return message.MessageID
}

func NewEditMarkUpMessage(userID int, markUp *tgbotapi.InlineKeyboardMarkup, text string) {
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      int64(userID),
			MessageID:   db.RdbGetAdminMsgID(userID),
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func SendAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, lang, text string) {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.LangText(lang, text),
	}

	if _, err := assets.Bot.AnswerCallbackQuery(answerCallback); err != nil {
		log.Println(err)
	}
}

func SendAdminAnswerCallback(callbackQuery *tgbotapi.CallbackQuery, text string) {
	lang := assets.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.AdminText(lang, text),
	}

	if _, err := assets.Bot.AnswerCallbackQuery(answerCallback); err != nil {
		log.Println(err)
	}
}

func GetFormatText(lang, text string, values ...interface{}) string {
	formatText := assets.LangText(lang, text)
	return fmt.Sprintf(formatText, values...)
}

func SendSimpleMsg(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
