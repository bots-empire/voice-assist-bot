package auth

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"time"
)

func (u *User) MakeMoney(botLang string) bool {
	if time.Now().Unix()/86400 > u.LastVoice/86400 {
		u.resetVoiceDayCounter(botLang)
	}

	if u.CompletedToday >= assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay {
		u.reachedMaxAmountPerDay(botLang)
		return false
	}

	db.RdbSetUser(botLang, u.ID, "make_money")

	u.sendMoneyStatistic(botLang)
	u.sendInvitationToRecord(botLang)
	return true
}

func (u *User) resetVoiceDayCounter(botLang string) {
	u.CompletedToday = 0
	u.LastVoice = time.Now().Unix()

	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		u.CompletedToday, u.LastVoice, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.40 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()
}

func (u *User) sendMoneyStatistic(botLang string) {
	text := assets.LangText(u.Language, "make_money_statistic")
	text = fmt.Sprintf(text, u.CompletedToday, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay,
		assets.AdminSettings.Parameters[botLang].VoiceAmount, u.Balance, u.CompletedToday*assets.AdminSettings.Parameters[botLang].VoiceAmount)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) sendInvitationToRecord(botLang string) {
	text := assets.LangText(u.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, assets.SiriText(u.Language))
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	back := tgbotapi.NewKeyboardButton(assets.LangText(u.Language, "back_to_main_menu_button"))
	row := tgbotapi.NewKeyboardButtonRow(back)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) reachedMaxAmountPerDay(botLang string) {
	text := assets.LangText(u.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay)
	msg := tgbotapi.NewMessage(int64(u.ID), text)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) AcceptVoiceMessage(botLang string) bool {
	u.Balance += assets.AdminSettings.Parameters[botLang].VoiceAmount
	u.Completed++
	u.CompletedToday++
	u.LastVoice = time.Now().Unix()

	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ?, last_voice = ? WHERE id = ?;",
		u.Balance, u.Completed, u.CompletedToday, u.LastVoice, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.89 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()

	return u.MakeMoney(botLang)
}

func (u *User) WithdrawMoneyFromBalance(botLang string, amount string) {
	amount = getAmountFromText(amount)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "incorrect_amount"))
		msgs2.SendMsgToUser(botLang, msg)
		return
	}

	if amountInt < assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount {
		u.minAmountNotReached(botLang)
		return
	}

	sendInvitationToSubs(botLang, u.ID, GetLang(botLang, u.ID), amount)
}

func getAmountFromText(text string) string {
	var amount string
	rs := []rune(text)
	for i := range rs {
		if rs[i] >= 48 && rs[i] <= 57 {
			amount += string(rs[i])
		}
	}
	return amount
}

func (u *User) minAmountNotReached(botLang string) {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	msgs2.NewParseMessage(botLang, int64(u.ID), text)
}

func sendInvitationToSubs(botLang string, userID int, userLang, amount string) {
	text := msgs2.GetFormatText(userLang, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[botLang].Url)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("im_subscribe_button", "withdrawal_exit/withdrawal_exit?"+amount)),
	).Build(userLang)

	msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) CheckSubscribeToWithdrawal(botLang string, callback *tgbotapi.CallbackQuery, userID, amount int) bool {
	if !u.CheckSubscribe(botLang, userID) {
		lang := GetLang(botLang, userID)
		msgs2.SendAnswerCallback(botLang, callback, lang, "user_dont_subscribe")
		return false
	}

	if u.Balance < amount {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "lack_of_funds"))
		msgs2.SendMsgToUser(botLang, msg)
		db.RdbSetUser(botLang, userID, "withdrawal/req_amount")
		return false
	}

	u.Balance -= amount
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ? WHERE id = ?;", u.Balance, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.163 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()

	msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "successfully_withdrawn"))
	msgs2.SendMsgToUser(botLang, msg)
	return true
}

func (u *User) CheckSubscribe(botLang string, userID int) bool {
	fmt.Println(assets.AdminSettings.AdvertisingChan[botLang].ChannelID)
	member, err := assets.Bots[botLang].Bot.GetChatMember(tgbotapi.ChatConfigWithUser{
		ChatID: assets.AdminSettings.AdvertisingChan[botLang].ChannelID,
		UserID: userID,
	})

	if err == nil {
		fmt.Println(member.Status)
		addMemberToSubsBase(botLang, userID)
		return checkMemberStatus(member)
	}
	fmt.Println(err, member)
	return false
}

func checkMemberStatus(member tgbotapi.ChatMember) bool {
	if member.IsAdministrator() {
		return true
	}
	if member.IsCreator() {
		return true
	}
	if member.IsMember() {
		return true
	}
	return false
}

func addMemberToSubsBase(botLang string, userId int) {
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM subs WHERE id = ?;", userId)
	if err != nil {
		text := "Fatal Err with DB - methods.207 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	user := readUser(rows)
	if user.ID != 0 {
		return
	}
	rows.Close()
	rows, err = dataBase.Query("INSERT INTO subs VALUES(?);", userId)
	if err != nil {
		text := "Fatal Err with DB - methods.219 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()
}

func readUser(rows *sql.Rows) User {
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var users []User

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			ID: id,
		})
	}
	if len(users) == 0 {
		users = append(users, User{
			ID: 0,
		})
	}
	return users[0]
}

func (u User) GetABonus(botLang string, callback *tgbotapi.CallbackQuery) {
	if !u.CheckSubscribe(botLang, u.ID) {
		lang := GetLang(botLang, u.ID)
		msgs2.SendAnswerCallback(botLang, callback, lang, "user_dont_subscribe")
		return
	}

	if u.TakeBonus {
		text := assets.LangText(u.Language, "bonus_already_have")
		msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
		return
	}

	u.Balance += assets.AdminSettings.Parameters[botLang].BonusAmount
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;", u.Balance, true, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.265 //" + err.Error()
		msgs2.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()

	text := assets.LangText(u.Language, "bonus_have_received")
	msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
}
