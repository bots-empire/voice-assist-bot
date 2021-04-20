package auth

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

func (user *User) WithdrawMoneyFromBalance(amount string) bool {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(user.ID), assets.LangText(user.Language, "incorrect_amount"))
		if _, err = assets.Bot.Send(msg); err != nil {
			log.Println(err)
		}
		return false
	}

	if amountInt < assets.AdminSettings.MinWithdrawalAmount {
		user.minAmountNotReached()
		return false
	}

	if user.Balance < amountInt {
		msg := tgbotapi.NewMessage(int64(user.ID), assets.LangText(user.Language, "lack_of_funds"))
		if _, err = assets.Bot.Send(msg); err != nil {
			log.Println(err)
		}
		return false
	}

	user.Balance -= amountInt
	_, err = db.DataBase.Query("UPDATE users SET balance = ? WHERE id = ?;", user.Balance, user.ID)
	if err != nil {
		panic(err.Error())
	}

	msg := tgbotapi.NewMessage(int64(user.ID), assets.LangText(user.Language, "successfully_withdrawn"))
	if _, err = assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
	return true
}

func (user *User) minAmountNotReached() {
	text := assets.LangText(user.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.MinWithdrawalAmount)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: int64(user.ID),
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
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
