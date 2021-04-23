package msgs

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func NewParseMessage(chatID int64, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: chatID,
		},
		Text:      text,
		ParseMode: "HTML",
	}
}

func NewParseMarkUpMessage(chatID int64, markUp tgbotapi.InlineKeyboardMarkup, text string) tgbotapi.MessageConfig {
	return tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      chatID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
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

func GetFormatText(lang, text string, values ...interface{}) string {
	formatText := assets.LangText(lang, text)
	return fmt.Sprintf(formatText, values...)
}
