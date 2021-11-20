package auth

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"time"

	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateBalanceQuery = "UPDATE users SET balance = ? WHERE id = ?;"

	updateAfterBonusQuery = "UPDATE users SET balance = ?, take_bonus = ? WHERE id = ?;"

	getSubsUserQuery = "SELECT * FROM subs WHERE id = ?;"
	updateSubsQuery  = "INSERT INTO subs VALUES(?);"
)

func WithdrawMoneyFromBalance(s model.Situation, amount string) error {
	amount = strings.Replace(amount, " ", "", -1)
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		msg := tgbotapi.NewMessage(int64(s.User.ID), assets.LangText(s.User.Language, "incorrect_amount"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	if amountInt < assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount {
		return minAmountNotReached(s.User, s.BotLang)
	}

	if s.User.Balance < amountInt {
		msg := tgbotapi.NewMessage(int64(s.User.ID), assets.LangText(s.User.Language, "lack_of_funds"))
		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	return sendInvitationToSubs(s, amount)
}

func minAmountNotReached(u *model.User, botLang string) error {
	text := assets.LangText(u.Language, "minimum_amount_not_reached")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount)

	return msgs.NewParseMessage(botLang, int64(u.ID), text)
}

func sendInvitationToSubs(s model.Situation, amount string) error {
	text := msgs.GetFormatText(s.User.Language, "withdrawal_not_subs_text")

	msg := tgbotapi.NewMessage(int64(s.User.ID), text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("im_subscribe_button", "/withdrawal_money?"+amount)),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func CheckSubscribeToWithdrawal(s model.Situation, amount int) bool {
	if s.User.Balance < amount {
		return false
	}

	if !CheckSubscribe(s) {
		_ = sendInvitationToSubs(s, strconv.Itoa(amount))
		return false
	}

	s.User.Balance -= amount
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateBalanceQuery, s.User.Balance, s.User.ID)
	if err != nil {
		return false
	}
	rows.Close()

	msg := tgbotapi.NewMessage(int64(s.User.ID), assets.LangText(s.User.Language, "successfully_withdrawn"))
	_ = msgs.SendMsgToUser(s.BotLang, msg)
	return true
}

func GetABonus(s model.Situation) error {
	if !CheckSubscribe(s) {
		text := assets.LangText(s.User.Language, "user_dont_subscribe")
		return msgs.SendSimpleMsg(s.BotLang, int64(s.User.ID), text)
	}

	if s.User.TakeBonus {
		text := assets.LangText(s.User.Language, "bonus_already_have")
		return msgs.SendSimpleMsg(s.BotLang, int64(s.User.ID), text)
	}

	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].BonusAmount
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(updateAfterBonusQuery, s.User.Balance, true, s.User.ID)
	if err != nil {
		return err
	}
	rows.Close()

	text := assets.LangText(s.User.Language, "bonus_have_received")
	return msgs.SendSimpleMsg(s.BotLang, int64(s.User.ID), text)
}

func CheckSubscribe(s model.Situation) bool {
	fmt.Println(assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID)
	member, err := model.Bots[s.BotLang].Bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: assets.AdminSettings.AdvertisingChan[s.BotLang].ChannelID,
			UserID: s.User.ID,
		}})

	if err == nil {
		if err := addMemberToSubsBase(s); err != nil {
			return false
		}
		return checkMemberStatus(member)
	}
	return false
}

func checkMemberStatus(member tgbotapi.ChatMember) bool {
	if member.IsAdministrator() {
		return true
	}
	if member.IsCreator() {
		return true
	}
	if member.Status == "member" {
		return true
	}
	return false
}

func addMemberToSubsBase(s model.Situation) error {
	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query(getSubsUserQuery, s.User.ID)
	if err != nil {
		return err
	}

	user, err := readUser(rows)
	if err != nil {
		return err
	}

	if user.ID != 0 {
		return nil
	}
	rows, err = dataBase.Query(updateSubsQuery, s.User.ID)
	if err != nil {
		return err
	}
	rows.Close()
	return nil
}

func readUser(rows *sql.Rows) (*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var id int64

		if err := rows.Scan(&id); err != nil {
			return nil, model.ErrScanSqlRow
		}

		users = append(users, &model.User{
			ID: id,
		})
	}
	if len(users) == 0 {
		users = append(users, &model.User{
			ID: 0,
		})
	}
	return users[0], nil
}

func AcceptVoiceMessage(s model.Situation) bool {
	s.User.Balance += assets.AdminSettings.Parameters[s.BotLang].VoiceAmount
	s.User.Completed++
	s.User.CompletedToday++
	s.User.LastVoice = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query("UPDATE users SET balance = ?, completed = ?, completed_today = ?, last_voice = ? WHERE id = ?;",
		s.User.Balance, s.User.Completed, s.User.CompletedToday, s.User.LastVoice, s.User.ID)
	if err != nil {
		text := "Fatal Err with DB - methods.89 //" + err.Error()
		_ = msgs.NewParseMessage("it", 1418862576, text)
		return false
	}
	err = rows.Close()
	if err != nil {
		return false
	}

	return MakeMoney(s)
}

func MakeMoney(s model.Situation) bool {
	var err error
	if time.Now().Unix()/86400 > s.User.LastVoice/86400 {
		err = resetVoiceDayCounter(s)
		if err != nil {
			return false
		}
	}

	if s.User.CompletedToday >= assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay {
		_ = reachedMaxAmountPerDay(s)
		return false
	}

	db.RdbSetUser(s.BotLang, s.User.ID, "/new_make_money")

	err = sendMoneyStatistic(s)
	if err != nil {
		return false
	}
	err = sendInvitationToRecord(s)
	if err != nil {
		return false
	}
	return true
}

func resetVoiceDayCounter(s model.Situation) error {
	s.User.CompletedToday = 0
	s.User.LastVoice = time.Now().Unix()

	dataBase := model.GetDB(s.BotLang)
	rows, err := dataBase.Query("UPDATE users SET completed_today = ?, last_voice = ? WHERE id = ?;",
		s.User.CompletedToday, s.User.LastVoice, s.User.ID)
	if err != nil {
		//msgs2.NewParseMessage("it", 1418862576, text)
		return fmt.Errorf("Fatal Err with DB - methods.40 //" + err.Error())
	}

	return rows.Close()
}

func sendMoneyStatistic(s model.Situation) error {
	text := assets.LangText(s.User.Language, "make_money_statistic")
	text = fmt.Sprintf(text, s.User.CompletedToday, assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay,
		assets.AdminSettings.Parameters[s.BotLang].VoiceAmount, s.User.Balance, s.User.CompletedToday*assets.AdminSettings.Parameters[s.BotLang].VoiceAmount)
	msg := tgbotapi.NewMessage(int64(s.User.ID), text)
	msg.ParseMode = "HTML"

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func sendInvitationToRecord(s model.Situation) error {
	text := assets.LangText(s.User.Language, "invitation_to_record_voice")
	text = fmt.Sprintf(text, assets.SiriText(s.User.Language))
	msg := tgbotapi.NewMessage(int64(s.User.ID), text)
	msg.ParseMode = "HTML"

	back := tgbotapi.NewKeyboardButton(assets.LangText(s.User.Language, "back_to_main_menu_button"))
	row := tgbotapi.NewKeyboardButtonRow(back)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func reachedMaxAmountPerDay(s model.Situation) error {
	text := assets.LangText(s.User.Language, "reached_max_amount_per_day")
	text = fmt.Sprintf(text, assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay, assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay)
	msg := tgbotapi.NewMessage(int64(s.User.ID), text)
	msg.ReplyMarkup = msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[s.User.Language].Url)),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
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
