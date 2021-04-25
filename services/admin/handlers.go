package admin

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	"github.com/Stepan1328/voice-assist-bot/services/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
	"strings"
)

func SetAdminLevel(message *tgbotapi.Message) {
	userID := message.From.ID
	if !containsInAdmin(userID) {
		notAdmin(userID)
		return
	}

	updateFirstNameInfo(message)
	db.DeleteOldAdminMsg(userID)

	setAdminBackButton(userID)
	sendAdminMainMenu(userID)
}

func containsInAdmin(userID int) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func notAdmin(userID int) {
	lang := auth.GetLang(userID)
	text := assets.LangText(lang, "not_admin")
	msgs.SendSimpleMsg(int64(userID), text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(userID int) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_log_in")

	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	msgs.NewParseMarkUpMessage(int64(userID), markUp, text)
}

func sendAdminMainMenu(userID int) {
	db.RdbSetUser(userID, "admin")
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_make_money_button", "admin/make_money")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_statistic_button", "admin/statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(userID) != 0 {
		msgs.NewEditMarkUpMessage(userID, &markUp, text)
		return
	}
	msgID := msgs.NewIDParseMarkUpMessage(int64(userID), markUp, text)
	db.RdbSetAdminMsgID(userID, msgID)
}

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
	if !checkBackButton(message, lang) {
		operationStatus(userID, lang, "operation_canceled")
		resendMakeMenuLevel(userID)
		return
	}

	newAmount, err := strconv.Atoi(message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		msgs.NewParseMessage(int64(userID), text)
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
	operationStatus(userID, lang, "operation_completed")
	resendMakeMenuLevel(userID)
}

func resendMakeMenuLevel(userID int) {
	db.DeleteOldAdminMsg(userID)

	db.RdbSetUser(userID, "admin/make_money")
	inlineMarkUp, text := sendMakeMoneyMenu(userID)
	msgID := msgs.NewIDParseMarkUpMessage(int64(userID), inlineMarkUp, text)
	db.RdbSetAdminMsgID(userID, msgID)
}

func checkBackButton(message *tgbotapi.Message, lang string) bool {
	backText := assets.AdminText(lang, "back_to_make_money_setting")
	if message.Text != backText {
		return true
	}
	return false
}

func operationStatus(userID int, lang, key string) {
	text := assets.AdminText(lang, key)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	msgs.NewParseMarkUpMessage(int64(userID), markUp, text)
}

// sendMenu is a local copy of global SendMenu
func sendMenu(userID int, text string) {
	db.RdbSetUser(userID, "main")

	msg := tgbotapi.NewMessage(int64(userID), text)
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("main_make_money")),
		msgs.NewRow(msgs.NewDataButton("main_profile"),
			msgs.NewDataButton("main_statistic")),
		msgs.NewRow(msgs.NewDataButton("main_withdrawal_of_money"),
			msgs.NewDataButton("main_money_for_a_friend")),
		msgs.NewRow(msgs.NewDataButton("main_more_money")),
	).Build(auth.GetLang(userID))

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}

func AnalyseAdminCallback(callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "admin/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "admin_setting":
		adminSettingCallbackLevel(callbackQuery)
	case "make_money":
		settingMakeMoneyCallbackLevel(callbackQuery)
	case "advertisement":
		msgs.SendAdminAnswerCallback(callbackQuery, "add_in_future")
	case "statistic":
		sendStatistic(callbackQuery.From.ID)
		msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
	}
}

func adminSettingCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeAdminSettingsCallbackLevel(callbackQuery)
		return
	}

	userID := callbackQuery.From.ID
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_stetting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("setting_language_button", "admin/admin_setting/language")),
		msgs.NewIlRow(msgs.NewIlAdminButton("admin_list_button", "admin/admin_setting/admin_list")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/admin_setting/back")),
	).Build(lang)

	msgs.NewEditMarkUpMessage(userID, &markUp, text)
	msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func analyzeAdminSettingsCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "admin_setting/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "language":
		changeAdminLanguage(callbackQuery)
	case "admin_list":
		msgs.SendAdminAnswerCallback(callbackQuery, "add_in_future")
	case "back":
		msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
		sendAdminMainMenu(callbackQuery.From.ID)
	}
}

func changeAdminLanguage(callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		setAdminLanguage(callbackQuery)
		return
	}

	userID := callbackQuery.From.ID
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_set_lang_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("set_lang_en", "admin/admin_setting/language/en"),
			msgs.NewIlAdminButton("set_lang_ru", "admin/admin_setting/language/ru")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting/language/back")),
	).Build(lang)

	msgs.NewEditMarkUpMessage(userID, &markUp, text)
	msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func setAdminLanguage(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := strings.Split(callbackQuery.Data, "/")[1]
	if lang != "back" {
		assets.AdminSettings.AdminID[userID].Language = lang
		assets.SaveAdminSettings()
	}

	callbackQuery.Data = "admin_setting"
	adminSettingCallbackLevel(callbackQuery)
}

func settingMakeMoneyCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeChangeParameterCallbackLevel(callbackQuery)
		return
	}

	userID := callbackQuery.From.ID
	markUp, text := sendMakeMoneyMenu(userID)
	msgs.NewEditMarkUpMessage(userID, markUp, text)
	msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func analyzeChangeParameterCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := assets.AdminLang(userID)
	var parameter string
	var value int

	callbackQuery.Data = strings.Replace(callbackQuery.Data, "make_money/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "bonus_amount":
		db.RdbSetUser(userID, "admin/make_money/bonus")
		parameter = assets.AdminText(lang, "change_bonus_amount_button")
		value = assets.AdminSettings.BonusAmount
	case "min_withdrawal_amount":
		db.RdbSetUser(userID, "admin/make_money/withdrawal")
		parameter = assets.AdminText(lang, "change_min_withdrawal_amount_button")
		value = assets.AdminSettings.MinWithdrawalAmount
	case "voice_amount":
		db.RdbSetUser(userID, "admin/make_money/voice")
		parameter = assets.AdminText(lang, "change_voice_amount_button")
		value = assets.AdminSettings.VoiceAmount
	case "voice_pd_amount":
		db.RdbSetUser(userID, "admin/make_money/voice_pd")
		parameter = assets.AdminText(lang, "change_voice_pd_amount_button")
		value = assets.AdminSettings.MaxOfVoicePerDay
	case "referral_amount":
		db.RdbSetUser(userID, "admin/make_money/referral")
		parameter = assets.AdminText(lang, "change_referral_amount_button")
		value = assets.AdminSettings.ReferralAmount
	case "back":
		level := db.GetLevel(userID)
		if strings.Count(level, "/") == 2 {
			operationStatus(userID, lang, "operation_canceled")
			resendMakeMenuLevel(userID)
			msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
			return
		}

		msgs.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
		sendAdminMainMenu(callbackQuery.From.ID)
		return
	}
	text := adminFormatText(lang, "set_new_amount_text", parameter, value)
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_make_money_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(lang)

	msgs.NewParseMarkUpMessage(int64(userID), markUp, text)
}

func sendMakeMoneyMenu(userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "make_money_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/make_money/bonus_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money/min_withdrawal_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_voice_amount_button", "admin/make_money/voice_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_voice_pd_amount_button", "admin/make_money/voice_pd_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/make_money/referral_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/make_money/back")),
	).Build(lang)

	db.RdbSetUser(userID, "admin/make_money")
	return &markUp, text
}

func settingAdvertisementCallbackLevel() {

}

func sendStatistic(userID int) {
	lang := assets.AdminLang(userID)

	assets.UploadAdminSettings()
	count := countUsers()
	blocked := countBlockedUsers()
	text := adminFormatText(lang, "statistic_text",
		count, blocked, count-blocked)

	msgs.NewParseMessage(int64(userID), text)
	db.DeleteOldAdminMsg(userID)
	sendAdminMainMenu(userID)
}

func adminFormatText(lang, key string, values ...interface{}) string {
	formatText := assets.AdminText(lang, key)
	return fmt.Sprintf(formatText, values...)
}

func simpleMsg(userID int, lang, key string) {
	text := assets.AdminText(lang, key)
	msg := tgbotapi.NewMessage(int64(userID), text)

	if _, err := assets.Bot.Send(msg); err != nil {
		log.Println(err)
	}
}
