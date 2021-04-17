package services

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
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
		if update.Message.Voice != nil {
			fmt.Println("It's gs")
		}
	}

	if update.Message != nil {
		//fmt.Println(update.Message)
		NewUser(update.Message)
		if update.Message.Command() == "start" {
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

	user.CheckingTheUser()
}

func checkMessage(message *tgbotapi.Message) {
	msgText := message.Text
	lang := auth.GetLang(message.From.ID)

	switch msgText {
	case assets.GetLangText(lang, "main_make_money"):
		MakeMoney(message)
	case assets.GetLangText(lang, "main_profile"):
		SendProfile(message)
	case assets.GetLangText(lang, "main_statistic"):
		SendStatistics(message)
	case assets.GetLangText(lang, "main_withdrawal_of_money"):
		WithdrawalMoney(message)
	case assets.GetLangText(lang, "main_money_for_a_friend"):
		SendReferralLink(message)
	case assets.GetLangText(lang, "main_more_money"):
		MoreMoney(message)
	case assets.GetLangText(lang, "main_back"):
		SendMenu(message, assets.GetLangText(lang, "back_to_main_menu"))
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

// SendProfile sends the user its statistics
func SendProfile(message *tgbotapi.Message) {
	user := auth.User{ID: message.From.ID}
	user.CheckingTheUser()

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
	text := assets.GetLangText(lang, "statistic_to_user")

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

// WithdrawalMoney performs money withdrawal // TODO:
func WithdrawalMoney(message *tgbotapi.Message) {
	text := "💰 <b>Balance:</b> %d $.\n\n" +
		"If you desire to withdraw funds, choose where you want to put them !"
	text = fmt.Sprintf(text, 150)

	channelURL := tgbotapi.NewInlineKeyboardButtonURL("📲 Go to the channel", "https://t.me/joinchat/Vm991h1lG-GNnoK6")
	row1 := tgbotapi.NewInlineKeyboardRow(channelURL)

	getBonuse := tgbotapi.NewInlineKeyboardButtonData("💰 Get bonus", "WithdrawalMoney/getBonuse")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonuse)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.MessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID:           message.Chat.ID,
			ReplyToMessageID: 0,
			ReplyMarkup:      markUp,
		},
		Text:                  text,
		ParseMode:             "HTML",
		DisableWebPagePreview: false,
	}

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

// SendReferralLink generates a referral link and sends it to the user // TODO:
func SendReferralLink(message *tgbotapi.Message) {
	text := "💼 Get bonuses inviting your friends\n" +
		"📲 Send a link to friends - https://t.me/anglvokale_bot?start=%d\n\n" +
		"15 $ - for each invited friend.\n\n\n" +
		"You have invited: %d (number of people)"
	text = fmt.Sprintf(text, message.From.ID, 0)

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

	getBonuse := tgbotapi.NewInlineKeyboardButtonData("💰 Get bonus", "MoreMoney/getBonuse")
	row2 := tgbotapi.NewInlineKeyboardRow(getBonuse)
	markUp := tgbotapi.NewInlineKeyboardMarkup(row1, row2)

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ReplyMarkup = markUp

	if _, err := Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
