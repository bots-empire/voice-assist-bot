package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"runtime/debug"
)

func panicCather() {
	msg := recover()
	if msg == nil {
		return
	}

	panicText := fmt.Sprintf("panic in backend: message = %s\n%s", msg, string(debug.Stack()))

	alertMsg := tgbotapi.NewMessage(1418862576, panicText)
	_ = msgs.SendMsgToUser("it", alertMsg)
}
