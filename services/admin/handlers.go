package admin

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
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

	setAdminBackButton(userID, "admin_log_in")
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
	msgs2.SendSimpleMsg(int64(userID), text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(userID int, key string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, key)

	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(int64(userID), markUp, text)
}

func sendAdminMainMenu(userID int) {
	db.RdbSetUser(userID, "admin")
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_make_money_button", "admin/make_money")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_statistic_button", "admin/statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(userID) != 0 {
		msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), &markUp, text)
		return
	}
	msgID := msgs2.NewIDParseMarkUpMessage(int64(userID), markUp, text)
	db.RdbSetAdminMsgID(userID, msgID)
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
		settingAdvertisementCallbackLevel(callbackQuery)
	case "statistic":
		sendStatistic(callbackQuery.From.ID)
		msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
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

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_language_button", "admin/admin_setting/language")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("admin_list_button", "admin/admin_setting/admin_list")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/admin_setting/back")),
	).Build(lang)

	msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), &markUp, text)
	msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func analyzeAdminSettingsCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "admin_setting/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "language":
		changeAdminLanguage(callbackQuery)
	case "admin_list":
		msgs2.SendAdminAnswerCallback(callbackQuery, "add_in_future")
	case "back":
		msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
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

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("set_lang_en", "admin/admin_setting/language/en"),
			msgs2.NewIlAdminButton("set_lang_ru", "admin/admin_setting/language/ru")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting/language/back")),
	).Build(lang)

	msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), &markUp, text)
	msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func setAdminLanguage(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := strings.Split(callbackQuery.Data, "/")[1]
	if lang != "back" {
		assets.AdminSettings.AdminID[userID].Language = lang
		assets.SaveAdminSettings()
	}

	setAdminBackButton(userID, "language_set")
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
	msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), markUp, text)
	msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
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
			setAdminBackButton(userID, "operation_canceled")
			resendMakeMenuLevel(userID)
			msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
			return
		}

		msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
		sendAdminMainMenu(callbackQuery.From.ID)
		return
	}

	msgs2.SendAdminAnswerCallback(callbackQuery, "type_the_text")
	text := adminFormatText(lang, "set_new_amount_text", parameter, value)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("back_to_make_money_setting")),
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(int64(userID), markUp, text)
}

func sendMakeMoneyMenu(userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "make_money_setting_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_bonus_amount_button", "admin/make_money/bonus_amount")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money/min_withdrawal_amount")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_voice_amount_button", "admin/make_money/voice_amount")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_voice_pd_amount_button", "admin/make_money/voice_pd_amount")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_referral_amount_button", "admin/make_money/referral_amount")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/make_money/back")),
	).Build(lang)

	db.RdbSetUser(userID, "admin/make_money")
	return &markUp, text
}

func settingAdvertisementCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeAdvertisementCallbackLevel(callbackQuery)
		return
	}

	userID := callbackQuery.From.ID
	markUp, text := sendAdvertisementMenu(userID)
	msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), markUp, text)
	msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func analyzeAdvertisementCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID

	callbackQuery.Data = strings.Replace(callbackQuery.Data, "advertisement/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "url":
		db.RdbSetUser(userID, "admin/advertisement/url")
		promptForInput(userID, "set_new_url_text", assets.AdminSettings.AdvertisingURL)
		msgs2.SendAdminAnswerCallback(callbackQuery, "type_the_text")
	case "change_text":
		analyzeChangeTextCallbackLevel(callbackQuery)
	case "mailing":
		analyzeMailingCallbackLevel(callbackQuery)
	case "back":
		msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
		sendAdminMainMenu(callbackQuery.From.ID)
	}
}

func promptForInput(userID int, key string, values ...interface{}) {
	lang := assets.AdminLang(userID)

	text := adminFormatText(lang, key, values...)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("back_to_advertisement_setting")),
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(int64(userID), markUp, text)
}

func analyzeChangeTextCallbackLevel(callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeLangOfChangeTextLevel(callbackQuery)
		return
	}

	db.RdbSetUser(callbackQuery.From.ID, "admin/advertisement/change_text")
	sendChangeTextMenu(callbackQuery.From.ID)
	msgs2.SendAdminAnswerCallback(callbackQuery, "make_a_choice")
}

func analyzeLangOfChangeTextLevel(callbackQuery *tgbotapi.CallbackQuery) {
	lang := strings.Replace(callbackQuery.Data, "change_text/", "", 1)
	switch lang {
	case "back":
		callbackQuery.Data = "advertisement"
		settingAdvertisementCallbackLevel(callbackQuery)
	default:
		userID := callbackQuery.From.ID
		db.RdbSetUser(userID, "admin/advertisement/change_text/"+lang)
		promptForInput(userID, "set_new_advertisement_text", assets.AdminSettings.AdvertisingText[lang])
		msgs2.SendAdminAnswerCallback(callbackQuery, "type_the_text")
	}
}

func sendChangeTextMenu(userID int) {
	lang := assets.AdminLang(userID)
	resetSelectedLang()

	text := assets.AdminText(lang, "change_text_of_advertisement_text")
	markUp := parseMainLanguageButton("change_text")
	markUp.Rows = append(markUp.Rows, msgs2.NewIlRow(
		msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement/change_text/back")),
	)
	replyMarkUp := markUp.Build(lang)

	if db.RdbGetAdminMsgID(userID) == 0 {
		msgID := msgs2.NewIDParseMarkUpMessage(int64(userID), &replyMarkUp, text)
		db.RdbSetAdminMsgID(userID, msgID)
		return
	}
	msgs2.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(userID), &replyMarkUp, text)
}

func parseMainLanguageButton(partition string) *msgs2.InlineMarkUp {
	markUp := msgs2.NewIlMarkUp()

	for _, lang := range assets.AvailableLang {
		button := "button_"
		if assets.AdminSettings.LangSelectedMap[lang] {
			button += "on_" + lang
		} else {
			button += "off_" + lang
		}
		markUp.Rows = append(markUp.Rows,
			msgs2.NewIlRow(msgs2.NewIlAdminButton(button, "admin/advertisement/"+partition+"/"+lang)),
		)
	}
	return &markUp
}

func sendAdvertisementMenu(userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "advertisement_setting_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_url_button", "admin/advertisement/url")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_text_button", "admin/advertisement/change_text")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("distribute_button", "admin/advertisement/mailing")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/advertisement/back")),
	).Build(lang)

	db.RdbSetUser(userID, "admin/advertisement")
	return &markUp, text
}

func sendStatistic(userID int) {
	lang := assets.AdminLang(userID)

	assets.UploadAdminSettings()
	count := countUsers()
	blocked := countBlockedUsers()
	text := adminFormatText(lang, "statistic_text",
		count, blocked, count-blocked)

	msgs2.NewParseMessage(int64(userID), text)
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
