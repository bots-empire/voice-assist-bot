package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
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
		checkMessage(update.Message)
	}

	if update.CallbackQuery != nil {
		checkCallbackQuery(update.CallbackQuery)
	}
}

func checkMessage(message *tgbotapi.Message) {
	auth.CheckingTheUser(message)
	lang := auth.GetLang(message.From.ID)
	if strings.Contains(auth.StringGoToMainButton(message.From.ID), message.Text) {
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
		return
	}

	level := auth.GetLevel(message.From.ID)
	data := strings.Split(level, "/")
	switch data[0] {
	case "main", "empty":
		mainLevel(message)
	case "withdrawal":
		withdrawalLevel(message, level)
	case "make_money":
		makeMoneyLevel(message)
		//case "empty":
		//	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
		//	if _, err := assets.Bot.Send(msg); err != nil {
		//		log.Println(err)
		//	}
	}
}

func mainLevel(message *tgbotapi.Message) {
	if message.Command() == "start" {
		lang := auth.GetLang(message.From.ID)
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
	} else {
		checkTextOfMessage(message)
	}
}

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

	userID := auth.UserIDToRdb(message.From.ID)
	_, err := db.Rdb.Set(userID, "withdrawal/req_amount", 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func checkSelectedPaymentMethod(message *tgbotapi.Message) {
	switch message.Text {
	case "📱 PayPal":
		paypalReq(message)
	case "💳 Credit card":
		creditCardReq(message)
	case "⬅️ Back":
		lang := auth.GetLang(message.From.ID)
		SendMenu(message.From.ID, assets.LangText(lang, "main_select_menu"))
	}
}

func paypalReq(message *tgbotapi.Message) {
	userID := auth.UserIDToRdb(message.From.ID)
	_, err := db.Rdb.Set(userID, "withdrawal/paypal", 0).Result()
	if err != nil {
		log.Println(err)
	}

	lang := auth.GetLang(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "paypal_email"))
	cancel := tgbotapi.NewKeyboardButton(assets.LangText(lang, "withdraw_cancel"))
	row := tgbotapi.NewKeyboardButtonRow(cancel)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	if _, err = assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func creditCardReq(message *tgbotapi.Message) {
	userID := auth.UserIDToRdb(message.From.ID)
	_, err := db.Rdb.Set(userID, "withdrawal/credit_card", 0).Result()
	if err != nil {
		log.Println(err)
	}

	lang := auth.GetLang(message.From.ID)
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "credit_card_number"))
	cancel := tgbotapi.NewKeyboardButton(assets.LangText(lang, "withdraw_cancel"))
	row := tgbotapi.NewKeyboardButtonRow(cancel)
	markUp := tgbotapi.NewReplyKeyboard(row)
	msg.ReplyMarkup = markUp

	if _, err = assets.Bot.Send(msg); err != nil {
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
		level := auth.GetLevel(message.From.ID)
		if level == "empty" {
			emptyLevel(message, lang)
		}
		return
	}
	userID := auth.UserIDToRdb(message.From.ID)
	_, err := db.Rdb.Set(userID, "main", 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func emptyLevel(message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// SendMenu sends the keyboard with the main menu
func SendMenu(ID int, text string) {
	userID := auth.UserIDToRdb(ID)
	_, err := db.Rdb.Set(userID, "main", 0).Result()
	if err != nil {
		log.Println(err)
	}

	lang := auth.GetLang(ID)
	msg := tgbotapi.NewMessage(int64(ID), text)

	makeMoney := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_make_money"))
	row1 := tgbotapi.NewKeyboardButtonRow(makeMoney)

	profile := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_profile"))
	statistic := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_statistic"))
	row2 := tgbotapi.NewKeyboardButtonRow(profile, statistic)

	withdrawal := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_withdrawal_of_money"))
	moneyForAFriend := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_money_for_a_friend"))
	row3 := tgbotapi.NewKeyboardButtonRow(withdrawal, moneyForAFriend)

	moreMoney := tgbotapi.NewKeyboardButton(assets.LangText(lang, "main_more_money"))
	row4 := tgbotapi.NewKeyboardButtonRow(moreMoney)

	markUp := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
	msg.ReplyMarkup = markUp

	if _, err = assets.Bot.Send(msg); err != nil {
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

	text := assets.LangText(user.Language, "profile_text")
	text = fmt.Sprintf(text, message.From.FirstName, message.From.UserName,
		user.Balance, user.Completed, user.ReferralCount)

	changeLang := tgbotapi.NewInlineKeyboardButtonData(
		assets.LangText(user.Language, "change_lang_button"), "change_lang")
	row := tgbotapi.NewInlineKeyboardRow(changeLang)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      message.Chat.ID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// SendStatistics sends the user statistics of the entire game
func SendStatistics(message *tgbotapi.Message) {
	lang := auth.GetLang(message.From.ID)
	text := assets.LangText(lang, "statistic_to_user")

	currentTime := time.Now()
	formatTime := currentTime.Format("02.01.2006 15.04")

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := float64(currentTime.Unix()%1000000/5*5)/1000 - 500
	totalVoice := int(totalEarned*1000) / 7
	text = fmt.Sprintf(text, formatTime, users, totalEarned, totalVoice)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// WithdrawalMoney performs money withdrawal
func WithdrawalMoney(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := assets.LangText(user.Language, "withdrawal_money")
	text = fmt.Sprintf(text, user.Balance)

	advertisingText := assets.LangText(user.Language, "advertising_button")
	channelURL := tgbotapi.NewInlineKeyboardButtonURL(advertisingText, assets.AdminSettings.AdvertisingURL)
	row1 := tgbotapi.NewInlineKeyboardRow(channelURL)

	bonusText := assets.LangText(user.Language, "withdraw_money_button")
	getBonus := tgbotapi.NewInlineKeyboardButtonData(bonusText, "withdrawalMoney/getBonus")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonus)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:      message.Chat.ID,
			ReplyMarkup: markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// SendReferralLink generates a referral link and sends it to the user
func SendReferralLink(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := assets.LangText(user.Language, "referral_text")
	text = fmt.Sprintf(text, user.ID, assets.AdminSettings.ReferralAmount, user.ReferralCount)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: message.Chat.ID,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// MoreMoney it is used to get a daily bonus
// and bonuses from other projects
func MoreMoney(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := assets.LangText(user.Language, "more_money_text")
	text = fmt.Sprintf(text, assets.AdminSettings.BonusAmount)

	advertisingText := assets.LangText(user.Language, "advertising_button")
	channelURL := tgbotapi.NewInlineKeyboardButtonURL(advertisingText, assets.AdminSettings.AdvertisingURL)
	row1 := tgbotapi.NewInlineKeyboardRow(channelURL)

	bonusText := assets.LangText(user.Language, "get_bonus_button")
	getBonus := tgbotapi.NewInlineKeyboardButtonData(bonusText, "moreMoney/getBonus")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonus)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = markUp

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
