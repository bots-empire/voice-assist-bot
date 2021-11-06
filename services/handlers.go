package services

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/admin"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	updatePrintHeader = "update number: %d	// voice][-bot-update:	"
	extraneousUpdate  = "extraneous update"
)

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel) {
	file, err := os.Create("assets/logs/" + strconv.FormatInt(time.Now().Unix(), 10) + "_" + botLang + ".txt")
	if err != nil {
		log.Println("err create file; " + err.Error())
		return
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println("err close file; " + err.Error())
		}
	}()

	for update := range updates {
		go checkUpdate(botLang, &update, file)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update, logFile *os.File) {
	defer panicCather()

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil {
		if update.Message.PinnedMessage != nil {
			return
		}
	}
	PrintNewUpdate(botLang, update)
	if update.Message != nil {
		checkMessage(botLang, update.Message, logFile)
		return
	}

	if update.CallbackQuery != nil {
		checkCallbackQuery(botLang, update.CallbackQuery, logFile)
		return
	}
}

func PrintNewUpdate(botLang string, update *tgbotapi.Update) {
	assets.UpdateStatistic.Mu.Lock()
	defer assets.UpdateStatistic.Mu.Unlock()

	if (time.Now().Unix())/86400 > int64(assets.UpdateStatistic.Day) {
		sendTodayUpdateMsg()
	}

	assets.UpdateStatistic.Counter++
	assets.SaveUpdateStatistic()

	fmt.Printf(updatePrintHeader, assets.UpdateStatistic.Counter)
	if update.Message != nil {
		if update.Message.Text != "" {
			fmt.Println(botLang, update.Message.Text)
			return
		}
	}

	if update.CallbackQuery != nil {
		fmt.Println(botLang, update.CallbackQuery.Data)
		return
	}

	fmt.Println(botLang, extraneousUpdate)
}

func sendTodayUpdateMsg() {
	text := "Today Update's counter: " + strconv.Itoa(assets.UpdateStatistic.Counter)
	msgID, _ := msgs2.NewIDParseMessage("it", 1418862576, text)
	_ = msgs2.SendMsgToUser("it", tgbotapi.PinChatMessageConfig{
		ChatID:    1418862576,
		MessageID: msgID,
	})
	assets.UpdateStatistic.Counter = 0
	assets.UpdateStatistic.Day = int(time.Now().Unix()) / 86400
}

func checkMessage(botLang string, message *tgbotapi.Message, logFile *os.File) {
	auth.CheckingUser(botLang, message)
	lang := auth.GetLang(botLang, message.From.ID)
	if message.Command() == "getUpdate" && message.From.ID == 1418862576 {
		text := "Now Update's counter: " + strconv.Itoa(assets.UpdateStatistic.Counter)
		_ = msgs2.NewParseMessage("it", 1418862576, text)
		return
	}

	if strings.Contains(auth.StringGoToMainButton(botLang, message.From.ID), message.Text) && message.Text != "" {
		if err := SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu")); err != nil {
			_, errWrite := logFile.WriteString(fmt.Sprintf(
				"[ERROR]; %s; err=%s\n", time.Now().Format("Jan _2 15:04:05.000000"), err.Error()))
			if errWrite != nil {
				log.Println(errWrite)
			}
			log.Println(err)
			smthWentWrong(botLang, message, lang)
		}
		return
	}

	if strings.Contains(message.Text, "new_admin") {
		s := admin.Situation{
			Message:  message,
			BotLang:  botLang,
			UserID:   message.From.ID,
			UserLang: auth.GetLang(botLang, message.From.ID),
			Command:  message.Text,
		}
		_ = admin.CheckNewAdmin(s)
		return
	}

	if message.Command() == "start" || message.Command() == "exit" {
		if err := SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu")); err != nil {
			_, errWrite := logFile.WriteString(fmt.Sprintf(
				"[ERROR]; %s; err=%s\n", time.Now().Format("Jan _2 15:04:05.000000"), err.Error()))
			if errWrite != nil {
				log.Println(errWrite)
			}
			log.Println(err)
			smthWentWrong(botLang, message, lang)
		}
		return
	}

	if message.Command() == "admin" {
		admin.SetAdminLevel(botLang, message)
		return
	}

	var err error
	level := db.GetLevel(botLang, message.From.ID)
	data := strings.Split(level, "/")
	switch data[0] {
	case "main", "empty":
		err = checkTextOfMessage(botLang, message)
	case "withdrawal":
		err = withdrawalLevel(botLang, message, level)
	case "make_money":
		err = makeMoneyLevel(botLang, message)
	case "admin":
		err = admin.AnalyzeAdminMessage(botLang, message, level)
	default:
		log.Println("default case")
		emptyLevel(botLang, message, lang)
	}

	if err != nil {
		_, errWrite := logFile.WriteString(fmt.Sprintf(
			"[ERROR]; %s; err = %s\n", time.Now().Format("Jan _2 15:04:05.000000"), err.Error()))
		if errWrite != nil {
			log.Println(errWrite)
		}
		log.Println(err)
		smthWentWrong(botLang, message, lang)
	}
}

func withdrawalLevel(botLang string, message *tgbotapi.Message, level string) error {
	data := strings.Split(level, "/")
	if len(data) == 1 {
		return checkSelectedPaymentMethod(botLang, message)
	}
	level = strings.Replace(level, "withdrawal/", "", 1)
	switch level {
	case "credit_card":
		if err := reqWithdrawalAmount(botLang, message); err != nil {
			return err
		}
	case "req_amount":
		user := auth.GetUser(botLang, message.From.ID)
		if err := user.WithdrawMoneyFromBalance(botLang, message.Text); err != nil {
			return err
		}
	}

	db.RdbSetUser(botLang, message.From.ID, "withdrawal/req_amount")
	return nil
}

func checkSelectedPaymentMethod(botLang string, message *tgbotapi.Message) error {
	lang := auth.GetLang(botLang, message.From.ID)
	switch message.Text {
	case assets.LangText(lang, "main_back"):
		return SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu"))
	default:
		return creditCardReq(botLang, message)
	}
}

func creditCardReq(botLang string, message *tgbotapi.Message) error {
	db.RdbSetUser(botLang, message.From.ID, "withdrawal/credit_card")

	lang := auth.GetLang(botLang, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "credit_card_number"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	return msgs2.SendMsgToUser(botLang, msg)
}

func reqWithdrawalAmount(botLang string, message *tgbotapi.Message) error {
	lang := auth.GetLang(botLang, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "req_withdrawal_amount"))

	return msgs2.SendMsgToUser(botLang, msg)
}

func makeMoneyLevel(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)
	if message.Voice == nil {
		msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(user.Language, "voice_not_recognized"))
		_ = msgs2.SendMsgToUser(botLang, msg)
		return nil
	}

	if !user.AcceptVoiceMessage(botLang) {
		return SendMenu(botLang, message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}
	return nil
}

func checkCallbackQuery(botLang string, callbackQuery *tgbotapi.CallbackQuery, logFile *os.File) {
	var err error
	switch strings.Split(callbackQuery.Data, "/")[0] {
	case "moreMoney":
		err = GetBonus(botLang, callbackQuery)
	case "withdrawal_exit":
		err = CheckSubsAndWithdrawal(botLang, callbackQuery, callbackQuery.From.ID)
	case "change_lang":
		err = ChangeLanguage(botLang, callbackQuery)
	case "admin":
		admin.AnalyseAdminCallback(botLang, callbackQuery)
	}

	if err != nil {
		_, errWrite := logFile.WriteString(fmt.Sprintf(
			"[ERROR]; %s; err = %s\n", time.Now().Format("Jan _2 15:04:05.000000"), err.Error()))
		if errWrite != nil {
			log.Println(errWrite)
		}
		log.Println(err)
	}
}

func checkTextOfMessage(botLang string, message *tgbotapi.Message) error {
	msgText := message.Text
	lang := auth.GetLang(botLang, message.From.ID)

	switch msgText {
	case assets.LangText(lang, "main_make_money"):
		return MakeMoney(botLang, message)
	case assets.LangText(lang, "main_profile"):
		return SendProfile(botLang, message)
	case assets.LangText(lang, "main_statistic"):
		return SendStatistics(botLang, message)
	case assets.LangText(lang, "main_withdrawal_of_money"):
		return WithdrawalMoney(botLang, message)
	case assets.LangText(lang, "main_money_for_a_friend"):
		return SendReferralLink(botLang, message)
	case assets.LangText(lang, "main_more_money"):
		return MoreMoney(botLang, message)
	default:
		level := db.GetLevel(botLang, message.From.ID)
		if level == "empty" || level == "main" {
			emptyLevel(botLang, message, lang)
			return nil
		}
	}

	return fmt.Errorf("msg not send to user")
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs2.SendMsgToUser(botLang, msg)
}

func smthWentWrong(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs2.SendMsgToUser(botLang, msg)
}

// SendMenu sends the keyboard with the main menu
func SendMenu(botLang string, userID int, text string) error {
	db.RdbSetUser(botLang, userID, "main")

	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("main_make_money")),
		msgs2.NewRow(msgs2.NewDataButton("main_profile"),
			msgs2.NewDataButton("main_statistic")),
		msgs2.NewRow(msgs2.NewDataButton("main_withdrawal_of_money"),
			msgs2.NewDataButton("main_money_for_a_friend")),
		msgs2.NewRow(msgs2.NewDataButton("main_more_money")),
	).Build(auth.GetLang(botLang, userID))

	return msgs2.SendMsgToUser(botLang, msg)
}

// MakeMoney allows you to earn money
// by accepting voice messages from the user
func MakeMoney(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)

	if !user.MakeMoney(botLang) {
		return SendMenu(botLang, message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}

	return nil
}

// SendProfile sends the user its statistics
func SendProfile(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "profile_text",
		message.From.FirstName, message.From.UserName, user.Balance, user.Completed, user.ReferralCount)

	return msgs2.NewParseMessage(botLang, int64(user.ID), text)
}

// SendStatistics sends the user statistics of the entire game
func SendStatistics(botLang string, message *tgbotapi.Message) error {
	lang := auth.GetLang(botLang, message.From.ID)
	text := assets.LangText(lang, "statistic_to_user")

	text = getDate(text)

	return msgs2.NewParseMessage(botLang, message.Chat.ID, text)
}

// WithdrawalMoney performs money withdrawal
func WithdrawalMoney(botLang string, message *tgbotapi.Message) error {
	db.RdbSetUser(botLang, message.From.ID, "withdrawal")

	return sendPaymentMethod(botLang, message)
}

// SendReferralLink generates a referral link and sends it to the user
func SendReferralLink(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "referral_text", cfg.GetBotConfig(botLang).Link,
		user.ID, assets.AdminSettings.Parameters[botLang].ReferralAmount, user.ReferralCount)

	return msgs2.NewParseMessage(botLang, message.Chat.ID, text)
}

// MoreMoney it is used to get a daily bonus
// and bonuses from other projects
func MoreMoney(botLang string, message *tgbotapi.Message) error {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "more_money_text",
		assets.AdminSettings.Parameters[botLang].BonusAmount, assets.AdminSettings.Parameters[botLang].BonusAmount)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[botLang].Url)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("get_bonus_button", "moreMoney/getBonus")),
	).Build(user.Language)

	return msgs2.SendMsgToUser(botLang, msg)
}

func getDate(text string) string {
	currentTime := time.Now()

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := currentTime.Unix() % 100000000 / 500 * 5
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text /*formatTime,*/, users, totalEarned, totalVoice)
}
