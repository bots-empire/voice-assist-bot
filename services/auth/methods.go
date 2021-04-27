package auth

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
	"time"
)

func (u *User) MakeMoney() bool {
	if time.Now().Unix()/86400 > u.LastVoice/86400 {
		u.resetVoiceDayCounter()
	}

	if u.CompletedToday >= assets.AdminSettings.MaxOfVoicePerDay {
		u.reachedMaxAmountPerDay()
		return false
	}

	db.RdbSetUser(u.ID, "make_money")

	u.sendMoneyStatistic()
	u.sendInvitationToRecord()
	return true
}

func (u *User) resetVoiceDayCounter() {
	u.CompletedToday = 0
	u.LastVoice = time.Now().Unix()

	_, err := db.DataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		u.CompletedToday, u.LastVoice, u.ID)
	if err != nil {
		panic(err.Error())
	}
}

func (u *User) sendMoneyStatistic() {
	text := assets.LangText(u.Language, "make_money_statistic")
	text = fmt.Sprintf(text, u.CompletedToday, assets.AdminSettings.MaxOfVoicePerDay,
		assets.AdminSettings.VoiceAmount, u.Balance, u.CompletedToday*assets.AdminSettings.VoiceAmount)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (u *User) sendInvitationToRecord() {
	text := assets.LangText(u.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, assets.SiriText(u.Language))
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	back := tgbotapi.NewKeyboardButton(assets.LangText(u.Language, "back_to_main_menu_button"))
	row := tgbotapi.NewKeyboardButtonRow(back)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (u *User) reachedMaxAmountPerDay() {
	text := assets.LangText(u.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.MaxOfVoicePerDay, assets.AdminSettings.MaxOfVoicePerDay)
	msg := tgbotapi.NewMessage(int64(u.ID), text)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (u *User) AcceptVoiceMessage() bool {
	u.Balance += assets.AdminSettings.VoiceAmount
	u.Completed++
	u.CompletedToday++
	u.LastVoice = time.Now().Unix()

	_, err := db.DataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ?, last_voice = ? WHERE id = ?;",
		u.Balance, u.Completed, u.CompletedToday, u.LastVoice, u.ID)
	if err != nil {
		panic(err.Error())
	}

	return u.MakeMoney()
}

func (u *User) WithdrawMoneyFromBalance(amount string) bool {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "incorrect_amount"))
		if _, err = assets.Bot.Send(msg); err != nil {
			log.Println(err)
		}
		return false
	}

	if amountInt < assets.AdminSettings.MinWithdrawalAmount {
		u.minAmountNotReached()
		return false
	}

	if u.Balance < amountInt {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "lack_of_funds"))
		if _, err = assets.Bot.Send(msg); err != nil {
			log.Println(err)
		}
		return false
	}

	u.Balance -= amountInt
	_, err = db.DataBase.Query("UPDATE users SET balance = ? WHERE id = ?;", u.Balance, u.ID)
	if err != nil {
		panic(err.Error())
	}

	msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "successfully_withdrawn"))
	if _, err = assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
	return true
}

func (u *User) minAmountNotReached() {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.MinWithdrawalAmount)

	msgs2.NewParseMessage(int64(u.ID), text)
}

func (u User) GetABonus() {
	if u.TakeBonus {
		text := assets.LangText(u.Language, "bonus_already_have")
		msgs2.SendSimpleMsg(int64(u.ID), text)
		return
	}

	u.Balance += assets.AdminSettings.BonusAmount
	_, err := db.DataBase.Query("UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;", u.Balance, true, u.ID)
	if err != nil {
		panic(err.Error())
	}

	text := assets.LangText(u.Language, "bonus_have_received")
	msgs2.SendSimpleMsg(int64(u.ID), text)
}
