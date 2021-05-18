package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/admin"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
	"time"
)

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		checkUpdate(botLang, &update)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update) {
	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil {
		checkMessage(botLang, update.Message)
		return
	}

	if update.CallbackQuery != nil {
		checkCallbackQuery(botLang, update.CallbackQuery)
		return
	}
}

func checkMessage(botLang string, message *tgbotapi.Message) {
	auth.CheckingTheUser(botLang, message)
	lang := auth.GetLang(botLang, message.From.ID)
	if strings.Contains(auth.StringGoToMainButton(botLang, message.From.ID), message.Text) && message.Text != "" {
		SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu"))
		return
	}

	if message.Command() == "start" || message.Command() == "exit" {
		SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu"))
		return
	}

	if message.Command() == "admin" {
		admin.SetAdminLevel(botLang, message)
		return
	}

	level := db.GetLevel(botLang, message.From.ID)
	data := strings.Split(level, "/")
	switch data[0] {
	case "main", "empty":
		checkTextOfMessage(botLang, message)
	case "withdrawal":
		withdrawalLevel(botLang, message, level)
	case "make_money":
		makeMoneyLevel(botLang, message)
	case "admin":
		admin.AnalyzeAdminMessage(botLang, message, level)
	default:
		emptyLevel(botLang, message, lang)
	}
}

func withdrawalLevel(botLang string, message *tgbotapi.Message, level string) {
	data := strings.Split(level, "/")
	if len(data) == 1 {
		checkSelectedPaymentMethod(botLang, message)
		return
	}
	level = strings.Replace(level, "withdrawal/", "", 1)
	switch level {
	case "paypal":
		reqWithdrawalAmount(botLang, message)
	case "credit_card":
		reqWithdrawalAmount(botLang, message)
	case "req_amount":
		user := auth.GetUser(botLang, message.From.ID)
		if user.WithdrawMoneyFromBalance(botLang, message.Text) {
			SendMenu(botLang, message.From.ID, assets.LangText(user.Language, "main_select_menu"))
		}
		return
	}

	db.RdbSetUser(botLang, message.From.ID, "withdrawal/req_amount")
}

func checkSelectedPaymentMethod(botLang string, message *tgbotapi.Message) {
	lang := auth.GetLang(botLang, message.From.ID)
	switch message.Text {
	case assets.LangText(lang, "paypal_method"):
		paypalReq(botLang, message)
	case assets.LangText(lang, "credit_card_method"):
		creditCardReq(botLang, message)
	case assets.LangText(lang, "main_back"):
		SendMenu(botLang, message.From.ID, assets.LangText(lang, "main_select_menu"))
	}
}

func paypalReq(botLang string, message *tgbotapi.Message) {
	db.RdbSetUser(botLang, message.From.ID, "withdrawal/paypal")

	lang := auth.GetLang(botLang, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "paypal_email"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	msgs2.SendMsgToUser(botLang, msg)
}

func creditCardReq(botLang string, message *tgbotapi.Message) {
	db.RdbSetUser(botLang, message.From.ID, "withdrawal/credit_card")

	lang := auth.GetLang(botLang, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "credit_card_number"))
	msg.ReplyMarkup = msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewDataButton("withdraw_cancel")),
	).Build(lang)

	msgs2.SendMsgToUser(botLang, msg)
}

func reqWithdrawalAmount(botLang string, message *tgbotapi.Message) {
	lang := auth.GetLang(botLang, message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "req_withdrawal_amount"))

	msgs2.SendMsgToUser(botLang, msg)
}

func makeMoneyLevel(botLang string, message *tgbotapi.Message) {
	if message.Voice == nil {
		return
	}

	user := auth.GetUser(botLang, message.From.ID)
	if !user.AcceptVoiceMessage(botLang) {
		SendMenu(botLang, message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}
}

func checkCallbackQuery(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	switch strings.Split(callbackQuery.Data, "/")[0] {
	case "moreMoney":
		GetBonus(botLang, callbackQuery)
	case "withdrawalMoney":
		Withdrawal(botLang, callbackQuery)
	case "change_lang":
		ChangeLanguage(botLang, callbackQuery)
	case "admin":
		admin.AnalyseAdminCallback(botLang, callbackQuery)
	}
}

func checkTextOfMessage(botLang string, message *tgbotapi.Message) {
	msgText := message.Text
	lang := auth.GetLang(botLang, message.From.ID)

	switch msgText {
	case assets.LangText(lang, "main_make_money"):
		MakeMoney(botLang, message)
	case assets.LangText(lang, "main_profile"):
		SendProfile(botLang, message)
	case assets.LangText(lang, "main_statistic"):
		SendStatistics(botLang, message)
	case assets.LangText(lang, "main_withdrawal_of_money"):
		WithdrawalMoney(botLang, message)
	case assets.LangText(lang, "main_money_for_a_friend"):
		SendReferralLink(botLang, message)
	case assets.LangText(lang, "main_more_money"):
		MoreMoney(botLang, message)
	default:
		level := db.GetLevel(botLang, message.From.ID)
		if level == "empty" || level == "main" {
			emptyLevel(botLang, message, lang)
		}
		return
	}
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	msgs2.SendMsgToUser(botLang, msg)
}

// SendMenu sends the keyboard with the main menu
func SendMenu(botLang string, userID int, text string) {
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

	msgs2.SendMsgToUser(botLang, msg)
}

// MakeMoney allows you to earn money
// by accepting voice messages from the user
func MakeMoney(botLang string, message *tgbotapi.Message) {
	user := auth.GetUser(botLang, message.From.ID)

	if !user.MakeMoney(botLang) {
		SendMenu(botLang, message.From.ID, assets.LangText(user.Language, "back_to_main_menu"))
	}
}

// SendProfile sends the user its statistics
func SendProfile(botLang string, message *tgbotapi.Message) {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "profile_text",
		message.From.FirstName, message.From.UserName, user.Balance, user.Completed, user.ReferralCount)

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlDataButton("change_lang_button", "change_lang")),
	).Build(user.Language)

	msgs2.NewParseMarkUpMessage(botLang, int64(user.ID), markUp, text)
}

// SendStatistics sends the user statistics of the entire game
func SendStatistics(botLang string, message *tgbotapi.Message) {
	lang := auth.GetLang(botLang, message.From.ID)
	text := assets.LangText(lang, "statistic_to_user")

	text = getDate(text)

	msgs2.NewParseMessage(botLang, message.Chat.ID, text)
}

// WithdrawalMoney performs money withdrawal
func WithdrawalMoney(botLang string, message *tgbotapi.Message) {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "withdrawal_money",
		user.Balance)

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingURL[user.Language])),
		msgs2.NewIlRow(msgs2.NewIlDataButton("withdraw_money_button", "withdrawalMoney/getBonus")),
	).Build(user.Language)

	msgs2.NewParseMarkUpMessage(botLang, int64(user.ID), markUp, text)
}

// SendReferralLink generates a referral link and sends it to the user
func SendReferralLink(botLang string, message *tgbotapi.Message) {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "referral_text", cfg.GetBotConfig(botLang).Link,
		user.ID, assets.AdminSettings.ReferralAmount, user.ReferralCount)

	msgs2.NewParseMessage(botLang, message.Chat.ID, text)
}

// MoreMoney it is used to get a daily bonus
// and bonuses from other projects
func MoreMoney(botLang string, message *tgbotapi.Message) {
	user := auth.GetUser(botLang, message.From.ID)

	text := msgs2.GetFormatText(user.Language, "more_money_text",
		assets.AdminSettings.BonusAmount)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingURL[user.Language])),
		msgs2.NewIlRow(msgs2.NewIlDataButton("get_bonus_button", "moreMoney/getBonus")),
	).Build(user.Language)

	msgs2.SendMsgToUser(botLang, msg)
}

func getDate(text string) string {
	currentTime := time.Now()
	formatTime := currentTime.Format("02.01.2006 15.04")

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := currentTime.Unix() % 100000000 / 500 * 5
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text, formatTime, users, totalEarned, totalVoice)
}
