package auth

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
	"time"
)

func (user *User) MakeMoney() bool {
	if time.Now().Unix()/86400 > user.LastVoice/86400 {
		user.resetVoiceDayCounter()
	}

	if user.CompletedToday >= assets.AdminSettings.MaxOfVoicePerDay {
		user.reachedMaxAmountPerDay()
		return false
	}

	userID := UserIDToRdb(user.ID)
	_, err := db.Rdb.Set(userID, "make_money", 0).Result()
	if err != nil {
		log.Println(err)
	}

	user.sendMoneyStatistic()
	user.sendInvitationToRecord()
	return true
}

func (user *User) resetVoiceDayCounter() {
	user.CompletedToday = 0
	user.LastVoice = time.Now().Unix()

	_, err := db.DataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		user.CompletedToday, user.LastVoice, user.ID)
	if err != nil {
		panic(err.Error())
	}
}

func (user *User) sendMoneyStatistic() {
	text := assets.LangText(user.Language, "make_money_statistic")
	text = fmt.Sprintf(text, user.CompletedToday, assets.AdminSettings.MaxOfVoicePerDay,
		assets.AdminSettings.VoiceAmount, user.Balance, user.CompletedToday*assets.AdminSettings.VoiceAmount)
	msg := tgbotapi.NewMessage(int64(user.ID), text)
	msg.ParseMode = "HTML"

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (user *User) sendInvitationToRecord() {
	text := assets.LangText(user.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, assets.SiriText(user.Language))
	msg := tgbotapi.NewMessage(int64(user.ID), text)
	msg.ParseMode = "HTML"

	back := tgbotapi.NewKeyboardButton(assets.LangText(user.Language, "back_to_main_menu_button"))
	row := tgbotapi.NewKeyboardButtonRow(back)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (user *User) reachedMaxAmountPerDay() {
	text := assets.LangText(user.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.MaxOfVoicePerDay, assets.AdminSettings.MaxOfVoicePerDay)
	msg := tgbotapi.NewMessage(int64(user.ID), text)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func (user *User) AcceptVoiceMessage() bool {
	user.Balance += assets.AdminSettings.VoiceAmount
	user.Completed++
	user.CompletedToday++
	user.LastVoice = time.Now().Unix()

	_, err := db.DataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ?, last_voice = ? WHERE id = ?;",
		user.Balance, user.Completed, user.CompletedToday, user.LastVoice, user.ID)
	if err != nil {
		panic(err.Error())
	}

	return user.MakeMoney()
}

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

func UserIDToRdb(id int) string {
	userID := "user:" + strconv.Itoa(id)
	return userID
}

func TemporaryIDToRdb(id int) string {
	temporaryID := "message:" + strconv.Itoa(id)
	return temporaryID
}
