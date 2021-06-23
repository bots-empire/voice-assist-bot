package msgs

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func SendMessageToChat(botLang string, msg tgbotapi.MessageConfig) bool {
	bot := assets.GetBot(botLang)
	if _, err := bot.Send(msg); err != nil {
		return false
	}
	return true
}

func NewParseMessage(botLang string, chatID int64, text string) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	SendMsgToUser(botLang, msg)
}

func NewIDParseMessage(botLang string, chatID int64, text string) int {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	bot := assets.GetBot(botLang)
	message, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return message.MessageID
}

func NewParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	SendMsgToUser(botLang, msg)
}

func NewIDParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) int {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	bot := assets.GetBot(botLang)
	message, err := bot.Send(msg)
	if err != nil {
		log.Println(err)
	}
	return message.MessageID
}

func NewEditMarkUpMessage(botLang string, userID, msgID int, markUp *tgbotapi.InlineKeyboardMarkup, text string) {
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      int64(userID),
			MessageID:   msgID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	SendMsgToUser(botLang, msg)
}

func SendAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, lang, text string) {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.LangText(lang, text),
	}

	SendAnswerCallbackToUser(botLang, answerCallback)
}

func SendAdminAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, text string) {
	lang := assets.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.AdminText(lang, text),
	}

	SendAnswerCallbackToUser(botLang, answerCallback)
}

func GetFormatText(lang, text string, values ...interface{}) string {
	formatText := assets.LangText(lang, text)
	return fmt.Sprintf(formatText, values...)
}

func SendSimpleMsg(botLang string, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	SendMsgToUser(botLang, msg)
}

func SendMsgToUser(botLang string, msg tgbotapi.Chattable) {
	bot := assets.GetBot(botLang)

	if _, err := bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func SendAnswerCallbackToUser(botLang string, callback tgbotapi.CallbackConfig) {
	bot := assets.GetBot(botLang)

	if _, err := bot.AnswerCallbackQuery(callback); err != nil {
		log.Println(err)
	}
}
