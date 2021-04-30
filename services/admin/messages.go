package admin

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

func AnalyzeAdminMessage(message *tgbotapi.Message, level string) {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	level = strings.Replace(level, "admin/", "", 1)
	data := strings.Split(level, "/")

	logOutText := assets.AdminText(lang, "exit")
	if message.Text == logOutText {
		db.DeleteOldAdminMsg(userID)
		simpleMsg(userID, lang, "admin_log_out")

		text := assets.LangText(lang, "main_select_menu")
		sendMenu(userID, text)
		return
	}

	switch data[0] {
	case "make_money":
		makeMoneyMessageLevel(message, level)
	case "advertisement":
		advertisementMessageLevel(message, level)
	}
}

func makeMoneyMessageLevel(message *tgbotapi.Message, level string) {
	if strings.Contains(level, "/") {
		level = strings.Replace(level, "make_money/", "", 1)
		changeMakeMoneySettingsLevel(message, level)
	}
}

func changeMakeMoneySettingsLevel(message *tgbotapi.Message, level string) {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	if !checkBackButton(message, lang, "back_to_make_money_setting") {
		setAdminBackButton(userID, "operation_canceled")
		resendMakeMenuLevel(userID)
		return
	}

	newAmount, err := strconv.Atoi(message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(int64(userID), text)
		return
	}

	switch level {
	case "bonus":
		assets.AdminSettings.BonusAmount = newAmount
	case "withdrawal":
		assets.AdminSettings.MinWithdrawalAmount = newAmount
	case "voice":
		assets.AdminSettings.VoiceAmount = newAmount
	case "voice_pd":
		assets.AdminSettings.MaxOfVoicePerDay = newAmount
	case "referral":
		assets.AdminSettings.ReferralAmount = newAmount
	}
	assets.SaveAdminSettings()
	setAdminBackButton(userID, "operation_completed")
	resendMakeMenuLevel(userID)
}

func resendMakeMenuLevel(userID int) {
	db.DeleteOldAdminMsg(userID)

	db.RdbSetUser(userID, "admin/make_money")
	inlineMarkUp, text := sendMakeMoneyMenu(userID)
	msgID := msgs2.NewIDParseMarkUpMessage(int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(userID, msgID)
}

func checkBackButton(message *tgbotapi.Message, lang, key string) bool {
	backText := assets.AdminText(lang, key)
	if message.Text != backText {
		return true
	}
	return false
}

func advertisementMessageLevel(message *tgbotapi.Message, level string) {
	if !strings.Contains(level, "/") {
		return
	}

	level = strings.Replace(level, "advertisement/", "", 1)
	data := strings.Split(level, "/")
	switch data[0] {
	case "change_url":
		changeAdvertisementTextLevel(message, level, "change_url")
	case "change_text":
		changeAdvertisementTextLevel(message, level, "change_text")
	}
}

//func changeUrlLevel(message *tgbotapi.Message) {
//	userID := message.From.ID
//	lang := assets.AdminLang(userID)
//	if !checkBackButton(message, lang, "back_to_advertisement_setting") {
//		setAdminBackButton(userID, "operation_canceled")
//		resendAdvertisementMenuLevel(userID)
//		return
//	}
//
//	if !regexp.InvitationLink.MatchString(message.Text) {
//		text := assets.AdminText(lang, "incorrect_url_change_input")
//		msgs2.NewParseMessage(int64(userID), text)
//		return
//	}
//
//	assets.AdminSettings.AdvertisingURL[lang] = message.Text
//	assets.SaveAdminSettings()
//	setAdminBackButton(userID, "operation_completed")
//	resendAdvertisementMenuLevel(userID)
//}

func resendAdvertisementMenuLevel(userID int) {
	db.DeleteOldAdminMsg(userID)

	db.RdbSetUser(userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(userID)
	msgID := msgs2.NewIDParseMarkUpMessage(int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(userID, msgID)
}

func changeAdvertisementTextLevel(message *tgbotapi.Message, level, capitation string) {
	if !strings.Contains(level, "/") {
		return
	}

	textLang := strings.Replace(level, capitation+"/", "", 1)
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	status := "operation_canceled"

	if checkBackButton(message, lang, "back_to_advertisement_setting") {
		switch capitation {
		case "change_url":
			assets.AdminSettings.AdvertisingURL[textLang] = message.Text
		case "change_text":
			assets.AdminSettings.AdvertisingText[textLang] = message.Text
		}
		assets.SaveAdminSettings()
		status = "operation_completed"
	}

	setAdminBackButton(userID, status)
	db.RdbSetUser(userID, "admin/advertisement/"+capitation)
	db.DeleteOldAdminMsg(userID)
	sendChangeWithLangMenu(userID, capitation)
}

// sendMenu is a local copy of global SendMenu
func sendMenu(userID int, text string) {
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
