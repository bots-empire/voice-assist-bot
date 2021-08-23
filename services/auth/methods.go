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
	var err error
	if time.Now().Unix()/86400 > u.LastVoice/86400 {
		err = u.resetVoiceDayCounter(botLang)
		if err != nil {
			return false
		}
	}

	if u.CompletedToday >= assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay {
		_ = u.reachedMaxAmountPerDay(botLang)
		return false
	}

	db.RdbSetUser(botLang, u.ID, "make_money")

	err = u.sendMoneyStatistic(botLang)
	if err != nil {
		return false
	}
	err = u.sendInvitationToRecord(botLang)
	if err != nil {
		return false
	}
	return true
}

func (u *User) resetVoiceDayCounter(botLang string) error {
	u.CompletedToday = 0
	u.LastVoice = time.Now().Unix()

	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		u.CompletedToday, u.LastVoice, u.ID)
	if err != nil {
		//msgs2.NewParseMessage("it", 1418862576, text)
		return fmt.Errorf("Fatal Err with DB - methods.40 //" + err.Error())
	}

	return rows.Close()
}

func (u *User) sendMoneyStatistic(botLang string) error {
	text := assets.LangText(u.Language, "make_money_statistic")
	text = fmt.Sprintf(text, u.CompletedToday, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay,
		assets.AdminSettings.Parameters[botLang].VoiceAmount, u.Balance, u.CompletedToday*assets.AdminSettings.Parameters[botLang].VoiceAmount)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	return msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) sendInvitationToRecord(botLang string) error {
	text := assets.LangText(u.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, assets.SiriText(u.Language))
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ParseMode = "HTML"

	back := tgbotapi.NewKeyboardButton(assets.LangText(u.Language, "back_to_main_menu_button"))
	row := tgbotapi.NewKeyboardButtonRow(back)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	return msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) reachedMaxAmountPerDay(botLang string) error {
	text := assets.LangText(u.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay, assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay)
	msg := tgbotapi.NewMessage(int64(u.ID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[u.Language].Url)),
	).Build(u.Language)

	return msgs2.SendMsgToUser(botLang, msg)
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
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return false
	}
	err = rows.Close()
	if err != nil {
		return false
	}

	return u.MakeMoney(botLang)
}

func (u *User) WithdrawMoneyFromBalance(botLang string, amount string) error {
	amount = getAmountFromText(amount)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "incorrect_amount"))
		return msgs2.SendMsgToUser(botLang, msg)
	}

	if amountInt < assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount {
		return u.minAmountNotReached(botLang)
	}

	return sendInvitationToSubs(botLang, u.ID, GetLang(botLang, u.ID), amount)
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

func (u *User) minAmountNotReached(botLang string) error {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	return msgs2.NewParseMessage(botLang, int64(u.ID), text)
}

func sendInvitationToSubs(botLang string, userID int, userLang, amount string) error {
	text := msgs2.GetFormatText(userLang, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[botLang].Url)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("im_subscribe_button", "withdrawal_exit/withdrawal_exit?"+amount)),
	).Build(userLang)

	return msgs2.SendMsgToUser(botLang, msg)
}

func (u *User) CheckSubscribeToWithdrawal(botLang string, callback *tgbotapi.CallbackQuery, userID, amount int) (bool, error) {
	member, err := u.CheckSubscribe(botLang, userID)
	if err != nil {
		return false, err
	}
	if !member {
		lang := GetLang(botLang, userID)
		err := msgs2.SendAnswerCallback(botLang, callback, lang, "user_dont_subscribe")
		return false, err
	}

	if u.Balance < amount {
		msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "lack_of_funds"))
		db.RdbSetUser(botLang, userID, "withdrawal/req_amount")
		err := msgs2.SendMsgToUser(botLang, msg)
		return false, err
	}

	u.Balance -= amount
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ? WHERE id = ?;", u.Balance, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.163 //" + err.Error()
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return false, err
	}
	err = rows.Close()
	if err != nil {
		return false, err
	}

	msg := tgbotapi.NewMessage(int64(u.ID), assets.LangText(u.Language, "successfully_withdrawn"))
	err = msgs2.SendMsgToUser(botLang, msg)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (u *User) CheckSubscribe(botLang string, userID int) (bool, error) {
	fmt.Println(assets.AdminSettings.AdvertisingChan[botLang].ChannelID)
	chatCfg := tgbotapi.ChatConfigWithUser{
		ChatID: assets.AdminSettings.AdvertisingChan[botLang].ChannelID,
		UserID: userID,
	}
	member, err := assets.Bots[botLang].Bot.GetChatMember(chatCfg)

	if err != nil {
		return false, err
	}

	fmt.Println(member.Status)
	err = addMemberToSubsBase(botLang, userID)
	if err != nil {
		return false, err
	}
	return checkMemberStatus(member), nil
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

func addMemberToSubsBase(botLang string, userId int) error {
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("SELECT * FROM subs WHERE id = ?;", userId)
	if err != nil {
		text := "Fatal Err with DB - methods.207 //" + err.Error()
		log.Println(text)
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return err
	}
	user := readUser(rows)
	if user.ID != 0 {
		return nil
	}
	rows, err = dataBase.Query("INSERT INTO subs VALUES(?);", userId)
	if err != nil {
		text := "Fatal Err with DB - methods.219 //" + err.Error()
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		log.Println(text)
		return err
	}

	return rows.Close()
}

func readUser(rows *sql.Rows) *User {
	defer func() {
		_ = rows.Close()
	}()

	var users []*User

	for rows.Next() {
		var id int

		if err := rows.Scan(&id); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, &User{
			ID: id,
		})
	}

	if len(users) == 0 {
		users = append(users, &User{
			ID: 0,
		})
	}
	return users[0]
}

func (u User) GetABonus(botLang string, callback *tgbotapi.CallbackQuery) error {
	member, err := u.CheckSubscribe(botLang, u.ID)
	if err != nil {
		return err
	}
	if !member {
		lang := GetLang(botLang, u.ID)
		return msgs2.SendAnswerCallback(botLang, callback, lang, "user_dont_subscribe")
	}

	if u.TakeBonus {
		text := assets.LangText(u.Language, "bonus_already_have")
		return msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
	}

	u.Balance += assets.AdminSettings.Parameters[botLang].BonusAmount
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;", u.Balance, true, u.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.265 //" + err.Error()
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return err
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	text := assets.LangText(u.Language, "bonus_have_received")
	return msgs2.SendSimpleMsg(botLang, int64(u.ID), text)
}
