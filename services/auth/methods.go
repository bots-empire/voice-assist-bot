package auth

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
)

func (user User) UpdateUserLevel(level string) {
	_, err := db.DataBase.Query("UPDATE users_level SET level = ? WHERE id = ?;", level, user.ID)
	if err != nil {
		panic(err.Error())
	}
}

func (user User) GetABonus() {
	if user.TakeBonus {
		text := assets.LangText(user.Language, "bonus_already_have")
		sendSimpleMsg(int64(user.ID), text)
		return
	}

	user.Balance += assets.AdminSettings.BonusAmount
	_, err := db.DataBase.Query("UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;", user.Balance, true, user.ID)
	if err != nil {
		panic(err.Error())
	}

	text := assets.LangText(user.Language, "bonus_have_received")
	sendSimpleMsg(int64(user.ID), text)
}

func sendSimpleMsg(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
