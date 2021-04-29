package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/admin"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strings"
	"time"
)

func ActionsWithUpdates(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		checkUpdate(&update)
	}
}

func checkUpdate(update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil {
		fmt.Println(update.Message.From.ID)
		checkMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		fmt.Println(update.CallbackQuery)
		checkCallbackQuery(update.CallbackQuery)
	}
}

func checkMessage(message *tgbotapi.Message) {
	auth.CheckingTheUser(message)
	lang := auth.GetLang(message.From.ID)
	if strings.Contains(auth.StringGoToMainButton(message.From.ID), message.Text) && message.Text != "" {
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
		return
	}

	if message.Command() == "start" || message.Command() == "exit" {
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
		return
	}

	if message.Command() == "admin" {
		admin.SetAdminLevel(message)
		return
	}

	level := db.GetLevel(message.From.ID)
	data := strings.Split(level, "/")
	switch data[0] {
	case "main", "empty":
		checkTextOfMessage(message)
	case "withdrawal":
		withdrawalLevel(message, level)
	case "make_money":
		makeMoneyLevel(message)
	case "admin":
		admin.AnalyzeAdminMessage(message, level)
	}
}

//func mainLevel(message *tgbotapi.Message) {
//	switch message.Command() {
//	case "start":
//		lang := auth.GetLang(message.From.ID)
//		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
//	default:
//		checkTextOfMessage(message)
//	}
//}

func withdrawalLevel(message *tgbotapi.Message, level string) {
	data := strings.Split(level, "/")
	if len(data) == 1 {
		checkSelectedPaymentMethod(message)
		return
	}
	level = strings.Replace(level, "withdrawal/", "", 1)
	switch level {
	case "paypal":
		reqWithdrawalAmount(message)
	case "credit_card":
		reqWithdrawalAmount(message)
	case "req_amount":
		user := auth.GetUser(message.From.ID)
		if user.WithdrawMoneyFromBalance(message.Text) {
			SendMenu(message.From.ID, assets.LangText(user.Language, "main_select_menu"))
		}
		return
	}

	db.RdbSetUser(message.From.ID, "withdrawal/req_amount")
}

func checkSelectedPaymentMethod(message *tgbotapi.Message) {
	lang := auth.GetLang(message.From.ID)
	switch message.Text {
	case assets.LangText(lang, "paypal_method"):
		paypalReq(message)
	case assets.LangText(lang, "credit_card_method"):
		creditCardReq(message)
	case assets.LangText(lang, "main_back"):
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
	}
}

func paypalReq(message *tgbotapi.Message) {
	db.RdbSetUser(message.From.ID, "withdrawal/paypal")

	lang := auth.GetLang(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "paypal_email"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func creditCardReq(message *tgbotapi.Message) {
	db.RdbSetUser(message.From.ID, "withdrawal/credit_card")

	lang := auth.GetLang(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "credit_card_number"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func reqWithdrawalAmount(message *tgbotapi.Message) {
	lang := auth.GetLang(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "req_withdrawal_amount"))

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func makeMoneyLevel(message *tgbotapi.Message) {
	if message.Voice == nil {
		return
	}

	user := auth.GetUser(message.From.ID)
	if !user.AcceptVoiceMessage() {
		SendMenu(message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}
}

func checkCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	switch strings.Split(callbackQuery.Data, "/")[0] {
	case "moreMoney":
		GetBonus(callbackQuery)
	case "withdrawalMoney":
		Withdrawal(callbackQuery)
	case "change_lang":
		ChangeLanguage(callbackQuery)
	case "admin":
		admin.AnalyseAdminCallback(callbackQuery)
	}
}

func checkTextOfMessage(message *tgbotapi.Message) {
	msgText := message.Text
	lang := auth.GetLang(message.From.ID)

	switch msgText {
	case assets.LangText(lang, "main_make_money"):
		MakeMoney(message)
	case assets.LangText(lang, "main_profile"):
		SendProfile(message)
	case assets.LangText(lang, "main_statistic"):
		SendStatistics(message)
	case assets.LangText(lang, "main_withdrawal_of_money"):
		WithdrawalMoney(message)
	case assets.LangText(lang, "main_money_for_a_friend"):
		SendReferralLink(message)
	case assets.LangText(lang, "main_more_money"):
		MoreMoney(message)
	default:
		level := db.GetLevel(message.From.ID)
		if level == "empty" {
			emptyLevel(message, lang)
		}
		return
	}
}

func emptyLevel(message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// SendMenu sends the keyboard with the main menu
func SendMenu(userID int, text string) {
	db.RdbSetUser(userID, "main")

	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("main_make_money")),
		msgs2.NewRow(msgs2.NewDataButton("main_profile"),
			msgs2.NewDataButton("main_statistic")),
		msgs2.NewRow(msgs2.NewDataButton("main_withdrawal_of_money"),
			msgs2.NewDataButton("main_money_for_a_friend")),
		msgs2.NewRow(msgs2.NewDataButton("main_more_money")),
	).Build(auth.GetLang(userID))

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// MakeMoney allows you to earn money
// by accepting voice messages from the user
func MakeMoney(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	if !user.MakeMoney() {
		SendMenu(message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}
}

// SendProfile sends the user its statistics
func SendProfile(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := msgs2.GetFormatText(user.Language, "profile_text",
		message.From.FirstName, message.From.UserName, user.Balance, user.Completed, user.ReferralCount)

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("change_lang_button", "change_lang")),
	).Build(user.Language)

	msgs2.NewParseMarkUpMessage(int64(user.ID), markUp, text)
}

// SendStatistics sends the user statistics of the entire game
func SendStatistics(message *tgbotapi.Message) {
	lang := auth.GetLang(message.From.ID)
	text := assets.LangText(lang, "statistic_to_user")

	text = getDate(text)

	msgs2.NewParseMessage(message.Chat.ID, text)
}

// WithdrawalMoney performs money withdrawal
func WithdrawalMoney(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := msgs2.GetFormatText(user.Language, "withdrawal_money",
		user.Balance)

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingURL)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("withdraw_money_button", "withdrawalMoney/getBonus")),
	).Build(user.Language)

	msgs2.NewParseMarkUpMessage(int64(user.ID), markUp, text)
}

// SendReferralLink generates a referral link and sends it to the user
func SendReferralLink(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := msgs2.GetFormatText(user.Language, "referral_text",
		user.ID, assets.AdminSettings.ReferralAmount, user.ReferralCount)

	msgs2.NewParseMessage(message.Chat.ID, text)
}

// MoreMoney it is used to get a daily bonus
// and bonuses from other projects
func MoreMoney(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := msgs2.GetFormatText(user.Language, "more_money_text",
		assets.AdminSettings.BonusAmount)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingURL)),
		msgs2.NewIlRow(msgs2.NewIlDataButton("get_bonus_button", "moreMoney/getBonus")),
	).Build(user.Language)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func getDate(text string) string {
	currentTime := time.Now()
	formatTime := currentTime.Format("02.01.2006 15.04")

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := currentTime.Unix()%1000000/5*5 - 500000
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text, formatTime, users, totalEarned, totalVoice)
}
