package admin

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

func AnalyzeAdminMessage(botLang string, message *tgbotapi.Message, level string) {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	level = strings.Replace(level, "admin/", "", 1)
	data := strings.Split(level, "/")

	logOutText := assets.AdminText(lang, "exit")
	if message.Text == logOutText {
		db.DeleteOldAdminMsg(botLang, userID)
		simpleMsg(botLang, userID, lang, "admin_log_out")

		text := assets.LangText(lang, "main_select_menu")
		sendMenu(botLang, userID, text)
		return
	}

	switch data[0] {
	case "make_money":
		makeMoneyMessageLevel(botLang, message, level)
	case "advertisement":
		advertisementMessageLevel(botLang, message, level)
	}
}

func makeMoneyMessageLevel(botLang string, message *tgbotapi.Message, level string) {
	if strings.Contains(level, "/") {
		level = strings.Replace(level, "make_money/", "", 1)
		changeMakeMoneySettingsLevel(botLang, message, level)
	}
}

func changeMakeMoneySettingsLevel(botLang string, message *tgbotapi.Message, level string) {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	if !checkBackButton(message, lang, "back_to_make_money_setting") {
		setAdminBackButton(botLang, userID, "operation_canceled")
		resendMakeMenuLevel(botLang, userID)
		return
	}

	newAmount, err := strconv.Atoi(message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs2.NewParseMessage(botLang, int64(userID), text)
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
	setAdminBackButton(botLang, userID, "operation_completed")
	resendMakeMenuLevel(botLang, userID)
}

func resendMakeMenuLevel(botLang string, userID int) {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/make_money")
	inlineMarkUp, text := sendMakeMoneyMenu(botLang, userID)
	msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(botLang, userID, msgID)
}

func checkBackButton(message *tgbotapi.Message, lang, key string) bool {
	backText := assets.AdminText(lang, key)
	if message.Text != backText {
		return true
	}
	return false
}

func advertisementMessageLevel(botLang string, message *tgbotapi.Message, level string) {
	if !strings.Contains(level, "/") {
		return
	}

	level = strings.Replace(level, "advertisement/", "", 1)
	data := strings.Split(level, "/")
	switch data[0] {
	case "change_url":
		changeAdvertisementTextLevel(botLang, message, level, "change_url")
	case "change_text":
		changeAdvertisementTextLevel(botLang, message, level, "change_text")
	}
}

func resendAdvertisementMenuLevel(botLang string, userID int) {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(botLang, userID)
	msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(botLang, userID, msgID)
}

func changeAdvertisementTextLevel(botLang string, message *tgbotapi.Message, level, capitation string) {
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
			assets.AdminSettings.AdvertisingChan[textLang].Url = message.Text
		case "change_text":
			assets.AdminSettings.AdvertisingText[textLang] = message.Text
		}
		assets.SaveAdminSettings()
		status = "operation_completed"
	}

	setAdminBackButton(botLang, userID, status)
	db.RdbSetUser(botLang, userID, "admin/advertisement/"+capitation)
	db.DeleteOldAdminMsg(botLang, userID)
	sendChangeWithLangMenu(botLang, userID, capitation)
}

// sendMenu is a local copy of global SendMenu
func sendMenu(botLang string, userID int, text string) {
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
