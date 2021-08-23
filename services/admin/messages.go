package admin

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strconv"
	"strings"
)

func AnalyzeAdminMessage(botLang string, message *tgbotapi.Message, level string) error {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	level = strings.Replace(level, "admin/", "", 1)
	data := strings.Split(level, "/")

	logOutText := assets.AdminText(lang, "exit")
	if message.Text == logOutText {
		db.DeleteOldAdminMsg(botLang, userID)
		simpleMsg(botLang, userID, lang, "admin_log_out")

		text := assets.LangText(lang, "main_select_menu")
		return sendMenu(botLang, userID, text)
	}

	switch data[0] {
	case "make_money":
		return makeMoneyMessageLevel(botLang, message, level)
	case "advertisement":
		return advertisementMessageLevel(botLang, message, level)
	case "delete_admin":
		return RemoveAdminCommand(botLang, message)
	}

	return nil
}

func makeMoneyMessageLevel(botLang string, message *tgbotapi.Message, level string) error {
	if strings.Contains(level, "/") {
		level = strings.Replace(level, "make_money/", "", 1)
		return changeMakeMoneySettingsLevel(botLang, message, level)
	}

	return nil
}

func changeMakeMoneySettingsLevel(botLang string, message *tgbotapi.Message, level string) error {
	userID := message.From.ID
	lang := assets.AdminLang(userID)
	if !checkBackButton(message, lang, "back_to_make_money_setting") {
		setAdminBackButton(botLang, userID, "operation_canceled")
		return resendMakeMenuLevel(botLang, userID)
	}

	newAmount, err := strconv.Atoi(message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs2.NewParseMessage(botLang, int64(userID), text)
	}

	switch level {
	case "bonus":
		assets.AdminSettings.Parameters[botLang].BonusAmount = newAmount
	case "withdrawal":
		assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount = newAmount
	case "voice":
		assets.AdminSettings.Parameters[botLang].VoiceAmount = newAmount
	case "voice_pd":
		assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay = newAmount
	case "referral":
		assets.AdminSettings.Parameters[botLang].ReferralAmount = newAmount
	}
	assets.SaveAdminSettings()
	setAdminBackButton(botLang, userID, "operation_completed")
	return resendMakeMenuLevel(botLang, userID)
}

func resendMakeMenuLevel(botLang string, userID int) error {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/make_money")
	inlineMarkUp, text := sendMakeMoneyMenu(botLang, userID)
	msgID, err := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), inlineMarkUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}

func checkBackButton(message *tgbotapi.Message, lang, key string) bool {
	backText := assets.AdminText(lang, key)
	if message.Text != backText {
		return true
	}
	return false
}

func advertisementMessageLevel(botLang string, message *tgbotapi.Message, level string) error {
	if !strings.Contains(level, "/") {
		return model.ErrSmthWentWrong
	}

	level = strings.Replace(level, "advertisement/", "", 1)
	data := strings.Split(level, "/")
	switch data[0] {
	case "change_url":
		return changeAdvertisementTextLevel(botLang, message, level, "change_url")
	case "change_text":
		return changeAdvertisementTextLevel(botLang, message, level, "change_text")
	}

	return model.ErrSmthWentWrong
}

func resendAdvertisementMenuLevel(botLang string, userID int) error {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := getAdvertisementMenu(botLang, userID)
	msgID, err := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), inlineMarkUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}

func changeAdvertisementTextLevel(botLang string, message *tgbotapi.Message, level, capitation string) error {
	if !strings.Contains(level, "/") {
		return model.ErrSmthWentWrong
	}

	userID := message.From.ID
	lang := assets.AdminLang(userID)
	status := "operation_canceled"

	if checkBackButton(message, lang, "back_to_advertisement_setting") {
		switch capitation {
		case "change_url":
			advertChan := getUrlAndChatID(message)
			if advertChan.ChannelID == 0 {
				text := assets.AdminText(lang, "chat_id_not_update")
				return msgs2.NewParseMessage(botLang, int64(userID), text)
			}

			assets.AdminSettings.AdvertisingChan[botLang] = advertChan
		case "change_text":
			assets.AdminSettings.AdvertisingText[botLang] = message.Text
		}
		assets.SaveAdminSettings()
		status = "operation_completed"
	}

	setAdminBackButton(botLang, userID, status)
	db.RdbSetUser(botLang, userID, "admin/advertisement/"+capitation)
	db.DeleteOldAdminMsg(botLang, userID)
	sendAdminMainMenu(botLang, userID)
	return nil
}

func getUrlAndChatID(message *tgbotapi.Message) *assets.AdvertChannel {
	data := strings.Split(message.Text, "\n")
	if len(data) != 2 {
		return &assets.AdvertChannel{}
	}

	chatId, err := strconv.Atoi(data[1])
	if err != nil {
		return &assets.AdvertChannel{}
	}

	return &assets.AdvertChannel{
		Url:       data[0],
		ChannelID: int64(chatId),
	}
}

// sendMenu is a local copy of global SendMenu
func sendMenu(botLang string, userID int, text string) error {
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
