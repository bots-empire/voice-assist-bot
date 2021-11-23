package services

import (
	"encoding/json"
	"fmt"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/administrator"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"runtime/debug"

	"github.com/Stepan1328/voice-assist-bot/log"
)

var (
	panicLogger = log.NewDefaultLogger().Prefix("panic cather")
)

func panicCather(botLang string, update *tgbotapi.Update) {
	msg := recover()
	if msg == nil {
		return
	}

	panicText := fmt.Sprintf("%s\npanic in backend: message = %s\n%s",
		botLang,
		msg,
		string(debug.Stack()),
	)
	panicLogger.Warn(panicText)

	alertMsg := tgbotapi.NewMessage(notificationChatID, panicText)
	_ = msgs2.SendMsgToUser(administrator.DefaultNotificationBot, alertMsg)

	data, err := json.MarshalIndent(update, "", "  ")
	if err != nil {
		return
	}

	updateDataMsg := tgbotapi.NewMessage(notificationChatID, string(data))
	_ = msgs2.SendMsgToUser(administrator.DefaultNotificationBot, updateDataMsg)
}
