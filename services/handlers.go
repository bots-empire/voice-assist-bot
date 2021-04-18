package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
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
		if update.Message.Voice != nil {
			fmt.Println("It's gs")
		}
	}

	if update.Message != nil {
		if update.Message.Command() == "start" {
			NewUser(update.Message)
			SendMenu(update.Message, "Select a menu item 👇")
		} else {
			checkMessage(update.Message)
		}
	}
}

func NewUser(message *tgbotapi.Message) {
	lang := message.From.LanguageCode
	if !strings.Contains("en,de,it,pt,es", lang) || lang == "" {
		lang = "en"
	}

	user := auth.User{
		ID:       message.From.ID,
		Language: lang,
	}

	referralID := PullReferralID(message)
	user.CheckingTheUser(referralID)
}

func PullReferralID(message *tgbotapi.Message) int {
	str := strings.Split(message.Text, " ")
	if len(str) < 2 {
		return 0
	}

	id, err := strconv.Atoi(str[1])
	if err != nil {
		log.Println(err)
		return 0
	}

	if id > 0 {
		return id
	}
	return 0
}

func checkMessage(message *tgbotapi.Message) {
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
	case assets.LangText(lang, "main_back"):
		SendMenu(message, assets.LangText(lang, "back_to_main_menu"))
	}
}

// SendMenu sends the keyboard with the main menu  // TODO:
func SendMenu(message *tgbotapi.Message, text string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)

	makeMoney := tgbotapi.NewKeyboardButton("👨🏻‍💻 Make money")
	row1 := tgbotapi.NewKeyboardButtonRow(makeMoney)

	profile := tgbotapi.NewKeyboardButton("👤 Profile")
	statistic := tgbotapi.NewKeyboardButton("📊 Statistics")
	row2 := tgbotapi.NewKeyboardButtonRow(profile, statistic)

	withdrawal := tgbotapi.NewKeyboardButton("💳 Withdrawal of money")
	moneyForAFriend := tgbotapi.NewKeyboardButton("💼 Money for a friend")
	row3 := tgbotapi.NewKeyboardButtonRow(withdrawal, moneyForAFriend)

	moreMoney := tgbotapi.NewKeyboardButton("💰 More money")
	row4 := tgbotapi.NewKeyboardButtonRow(moreMoney)

	markUp := tgbotapi.NewReplyKeyboard(row1, row2, row3, row4)
	msg.ReplyMarkup = markUp

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// MakeMoney allows you to earn money
// by accepting voice messages from the user // TODO:
func MakeMoney(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "✅ You have already sent 20/20 voice messages! "+
		"Come back in 24 hours to continue earning money...")

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}

	SendMenu(message, "You are back to the main menu")
}

// SendProfile sends the user its statistics // TODO:
func SendProfile(message *tgbotapi.Message) {
	user := auth.GetUser(message.From.ID)

	text := "👤 My profile:\n\n" +
		"<b>Name:</b> %s\n" +
		"<b>Username:</b> %s\n" +
		"<b>Balance:</b> %d $\n" +
		"<b>Voice messages that you have sent:</b> %d\n" +
		"<b>Invited:</b> %d" // получить исходя из языка

	text = fmt.Sprintf(text, message.From.FirstName, message.From.UserName,
		user.Balance, user.Completed, user.ReferralCount) //вставить баланс, количество гс и инвайтов

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           message.Chat.ID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	if _, err := Bot.Send(msg); err != nil {
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
			ChatID:           message.Chat.ID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	if _, err := Bot.Send(msg); err != nil {
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

	bonusText := assets.LangText(user.Language, "get_bonus_button")
	getBonus := tgbotapi.NewInlineKeyboardButtonData(bonusText, "WithdrawalMoney/getBonus")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonus)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           message.Chat.ID,
			ReplyToMessageID: 0,
			ReplyMarkup:      markUp,
		},
		Text:      text,
		ParseMode: "HTML",
	}

	if _, err := Bot.Send(msg); err != nil {
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
			ChatID:           message.Chat.ID,
			ReplyToMessageID: 0,
		},
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// MoreMoney it is used to get a daily bonus
// and bonuses from other projects // TODO:
func MoreMoney(message *tgbotapi.Message) {
	text := "Earn more !\n\n" +
		"To earn an extra £ 50, subscribe to the partner channel and watch the previous 15 posts !"

	channelURL := tgbotapi.NewInlineKeyboardButtonURL("📲 Go to the channel", "https://t.me/joinchat/Vm991h1lG-GNnoK6")
	row1 := tgbotapi.NewInlineKeyboardRow(channelURL)

	getBonus := tgbotapi.NewInlineKeyboardButtonData("💰 Get bonus", "MoreMoney/getBonus")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonus)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = markUp

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
