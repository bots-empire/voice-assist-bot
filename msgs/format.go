package msgs

import (
	"fmt"

	"github.com/Stepan1328/voice-assist-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func SendMessageToChat(botLang string, msg tgbotapi.MessageConfig) bool {
	bot := assets.GetBot(botLang)
	if _, err := bot.Send(msg); err != nil {
		return false
	}
	return true
}

func NewParseMessage(botLang string, chatID int64, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewIDParseMessage(botLang string, chatID int64, text string) (int, error) {
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
		return 0, err
	}
	return message.MessageID, nil
}

func NewParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) error {
	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func NewIDParseMarkUpMessage(botLang string, chatID int64, markUp interface{}, text string) (int, error) {
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
		return 0, err
	}
	return message.MessageID, nil
}

func NewEditMarkUpMessage(botLang string, userID, msgID int, markUp *tgbotapi.InlineKeyboardMarkup, text string) error {
	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      int64(userID),
			MessageID:   msgID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	return SendMsgToUser(botLang, msg)
}

func SendAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, lang, text string) error {
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.LangText(lang, text),
	}

	return SendAnswerCallbackToUser(botLang, answerCallback)
}

func SendAdminAnswerCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery, text string) error {
	lang := assets.AdminLang(callbackQuery.From.ID)
	answerCallback := tgbotapi.CallbackConfig{
		CallbackQueryID: callbackQuery.ID,
		Text:            assets.AdminText(lang, text),
	}

	return SendAnswerCallbackToUser(botLang, answerCallback)
}

func GetFormatText(lang, text string, values ...interface{}) string {
	formatText := assets.LangText(lang, text)
	return fmt.Sprintf(formatText, values...)
}

func SendSimpleMsg(botLang string, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)

	return SendMsgToUser(botLang, msg)
}

func SendMsgToUser(botLang string, msg tgbotapi.Chattable) error {
	bot := assets.GetBot(botLang)

	if _, err := bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func SendAnswerCallbackToUser(botLang string, callback tgbotapi.CallbackConfig) error {
	bot := assets.GetBot(botLang)

	if _, err := bot.AnswerCallbackQuery(callback); err != nil {
		return err
	}
	return nil
}
