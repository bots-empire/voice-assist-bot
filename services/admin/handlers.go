package admin

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func SetAdminLevel(botLang string, message *tgbotapi.Message) {
	userID := message.From.ID
	if !containsInAdmin(userID) {
		notAdmin(botLang, userID)
		return
	}

	updateFirstNameInfo(message)
	db.DeleteOldAdminMsg(botLang, userID)

	setAdminBackButton(botLang, userID, "admin_log_in")
	sendAdminMainMenu(botLang, userID)
}

func containsInAdmin(userID int) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

func notAdmin(botLang string, userID int) error {
	lang := auth.GetLang(botLang, userID)
	text := assets.LangText(lang, "not_admin")
	return msgs2.SendSimpleMsg(botLang, int64(userID), text)
}

func updateFirstNameInfo(message *tgbotapi.Message) {
	userID := message.From.ID
	assets.AdminSettings.AdminID[userID].FirstName = message.From.FirstName
	assets.SaveAdminSettings()
}

func setAdminBackButton(botLang string, userID int, key string) error {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, key)

	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	return msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
}

func sendAdminMainMenu(botLang string, userID int) error {
	db.RdbSetUser(botLang, userID, "admin")
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "admin_main_menu_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_admin_button", "admin/admin_setting")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_make_money_button", "admin/make_money")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_advertisement_button", "admin/advertisement")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("setting_statistic_button", "admin/statistic")),
	).Build(lang)

	if db.RdbGetAdminMsgID(botLang, userID) != 0 {
		return msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
	}
	msgID, err := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), markUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}

func AnalyseAdminCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "admin/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "admin_setting":
		adminSettingCallbackLevel(botLang, callbackQuery)
	case "make_money":
		settingMakeMoneyCallbackLevel(botLang, callbackQuery)
	case "advertisement":
		settingAdvertisementCallbackLevel(botLang, callbackQuery)
	case "statistic":
		sendStatistic(botLang, callbackQuery.From.ID)
		_ = msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
	}
}

func adminSettingCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeAdminSettingsCallbackLevel(botLang, callbackQuery)
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

	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func analyzeAdminSettingsCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "admin_setting/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "language":
		changeAdminLanguage(botLang, callbackQuery)
	case "admin_list":
		SendAdminListMenu(botLang, callbackQuery)
	case "back":
		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		sendAdminMainMenu(botLang, callbackQuery.From.ID)
	}
}

func changeAdminLanguage(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		setAdminLanguage(botLang, callbackQuery)
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

	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func setAdminLanguage(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := strings.Split(callbackQuery.Data, "/")[1]
	if lang != "back" {
		assets.AdminSettings.AdminID[userID].Language = lang
		assets.SaveAdminSettings()
	}

	setAdminBackButton(botLang, userID, "language_set")
	callbackQuery.Data = "admin_setting"
	adminSettingCallbackLevel(botLang, callbackQuery)
}

func settingMakeMoneyCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeChangeParameterCallbackLevel(botLang, callbackQuery)
		return
	}

	userID := callbackQuery.From.ID
	markUp, text := sendMakeMoneyMenu(botLang, userID)
	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), markUp, text)
	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func analyzeChangeParameterCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	lang := assets.AdminLang(userID)
	var parameter string
	var value int

	callbackQuery.Data = strings.Replace(callbackQuery.Data, "make_money/", "", 1)
	data := strings.Split(callbackQuery.Data, "/")
	switch data[0] {
	case "bonus_amount":
		db.RdbSetUser(botLang, userID, "admin/make_money/bonus")
		parameter = assets.AdminText(lang, "change_bonus_amount_button")
		value = assets.AdminSettings.Parameters[botLang].BonusAmount
	case "min_withdrawal_amount":
		db.RdbSetUser(botLang, userID, "admin/make_money/withdrawal")
		parameter = assets.AdminText(lang, "change_min_withdrawal_amount_button")
		value = assets.AdminSettings.Parameters[botLang].MinWithdrawalAmount
	case "voice_amount":
		db.RdbSetUser(botLang, userID, "admin/make_money/voice")
		parameter = assets.AdminText(lang, "change_voice_amount_button")
		value = assets.AdminSettings.Parameters[botLang].VoiceAmount
	case "voice_pd_amount":
		db.RdbSetUser(botLang, userID, "admin/make_money/voice_pd")
		parameter = assets.AdminText(lang, "change_voice_pd_amount_button")
		value = assets.AdminSettings.Parameters[botLang].MaxOfVoicePerDay
	case "referral_amount":
		db.RdbSetUser(botLang, userID, "admin/make_money/referral")
		parameter = assets.AdminText(lang, "change_referral_amount_button")
		value = assets.AdminSettings.Parameters[botLang].ReferralAmount
	case "back":
		level := db.GetLevel(botLang, userID)
		if strings.Count(level, "/") == 2 {
			setAdminBackButton(botLang, userID, "operation_canceled")
			db.DeleteOldAdminMsg(botLang, userID)
		}

		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		sendAdminMainMenu(botLang, userID)
		return
	}

	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "type_the_text")
	text := adminFormatText(lang, "set_new_amount_text", parameter, value)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("back_to_make_money_setting")),
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
}

func sendMakeMoneyMenu(botLang string, userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
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

	db.RdbSetUser(botLang, userID, "admin/make_money")
	return &markUp, text
}

func settingAdvertisementCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeAdvertisementCallbackLevel(botLang, callbackQuery)
		return
	}

	sendAdvertisementMenu(botLang, callbackQuery)
}

func sendAdvertisementMenu(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID
	msgID := db.RdbGetAdminMsgID(botLang, userID)
	markUp, text := getAdvertisementMenu(botLang, userID)
	if msgID == 0 {
		msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
	} else {
		msgs2.NewEditMarkUpMessage(botLang, userID, msgID, markUp, text)
	}

	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func analyzeAdvertisementCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	userID := callbackQuery.From.ID

	data := strings.Split(callbackQuery.Data, "/")
	switch data[1] {
	case "change_url":
		analyzeChangeUrlLevel(botLang, callbackQuery)
	case "change_text":
		analyzeChangeTextCallbackLevel(botLang, callbackQuery)
	case "mailing":
		analyzeMailingCallbackLevel(botLang, callbackQuery)
	case "back":
		level := db.GetLevel(botLang, userID)
		if strings.Count(level, "/") == 2 {
			setAdminBackButton(botLang, userID, "operation_canceled")
			db.DeleteOldAdminMsg(botLang, userID)
		}

		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		sendAdminMainMenu(botLang, userID)
	}
}

func analyzeChangeUrlLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	//if strings.Contains(callbackQuery.Data, "/") {
	analyzeLangOfChangeTextOrUrlLevel(botLang, callbackQuery, "change_url")
	//	return
	//}
	//
	//db.RdbSetUser(botLang, callbackQuery.From.ID, "admin/advertisement/change_url")
	//sendChangeWithLangMenu(botLang, callbackQuery.From.ID, "change_url")
	//msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func promptForInput(botLang string, userID int, key string, values ...interface{}) {
	lang := assets.AdminLang(userID)

	text := adminFormatText(lang, key, values...)
	markUp := msgs2.NewMarkUp(
		msgs2.NewRow(msgs2.NewAdminButton("back_to_advertisement_setting")),
		msgs2.NewRow(msgs2.NewAdminButton("exit")),
	).Build(lang)

	msgs2.NewParseMarkUpMessage(botLang, int64(userID), markUp, text)
}

func analyzeChangeTextCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	analyzeLangOfChangeTextOrUrlLevel(botLang, callbackQuery, "change_text")

	//if strings.Contains(callbackQuery.Data, "/") {
	//analyzeLangOfChangeTextOrUrlLevel(botLang, callbackQuery, "change_text")
	//return
	//}
	//
	//db.RdbSetUser(botLang, callbackQuery.From.ID, "admin/advertisement/change_text")
	//sendChangeWithLangMenu(botLang, callbackQuery.From.ID, "change_text")
	//msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func analyzeLangOfChangeTextOrUrlLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery, partition string) {
	userID := callbackQuery.From.ID
	lang := strings.Replace(callbackQuery.Data, "advertisement/"+partition+"/", "", 1)
	var key, value string
	switch partition {
	case "change_text":
		key = "set_new_advertisement_text"
		value = assets.AdminSettings.AdvertisingText[botLang]
	case "change_url":
		key = "set_new_url_text"
		value = assets.AdminSettings.AdvertisingChan[botLang].Url
	}

	switch lang {
	case "back":
		level := db.GetLevel(botLang, userID)
		if strings.Count(level, "/") == 3 {
			setAdminBackButton(botLang, userID, "operation_canceled")
			db.DeleteOldAdminMsg(botLang, userID)
		}

		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		callbackQuery.Data = "advertisement"
		settingAdvertisementCallbackLevel(botLang, callbackQuery)
	default:
		db.RdbSetUser(botLang, userID, "admin/advertisement/"+partition+"/"+lang)
		promptForInput(botLang, userID, key, value)
		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "type_the_text")
	}
}

func sendChangeWithLangMenu(botLang string, userID int, partition string) {
	lang := assets.AdminLang(userID)
	resetSelectedLang()
	key := partition + "_of_advertisement_text"

	text := assets.AdminText(lang, key)
	markUp := parseMainLanguageButton(partition)
	markUp.Rows = append(markUp.Rows, msgs2.NewIlRow(
		msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement/"+partition+"/back")),
	)
	replyMarkUp := markUp.Build(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID, _ := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), &replyMarkUp, text)
		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return
	}
	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &replyMarkUp, text)
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

func getAdvertisementMenu(botLang string, userID int) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := assets.AdminLang(userID)
	text := assets.AdminText(lang, "advertisement_setting_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_url_button", "admin/advertisement/change_url")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("change_text_button", "admin/advertisement/change_text")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("distribute_button", "admin/advertisement/mailing")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_main_menu", "admin/advertisement/back")),
	).Build(lang)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	return &markUp, text
}

func sendStatistic(botLang string, userID int) {
	lang := assets.AdminLang(userID)

	assets.UploadAdminSettings()
	allCount := countAllUsers()
	count := countUsers(botLang)
	referrals := countReferrals(botLang, count)
	blocked := countBlockedUsers(botLang)
	subscribers := countSubscribers(botLang)
	text := adminFormatText(lang, "statistic_text",
		allCount, count, referrals, blocked, subscribers, count-blocked)

	msgs2.NewParseMessage(botLang, int64(userID), text)
	db.DeleteOldAdminMsg(botLang, userID)
	sendAdminMainMenu(botLang, userID)
}

func adminFormatText(lang, key string, values ...interface{}) string {
	formatText := assets.AdminText(lang, key)
	return fmt.Sprintf(formatText, values...)
}

func simpleMsg(botLang string, userID int, lang, key string) {
	text := assets.AdminText(lang, key)
	msg := tgbotapi.NewMessage(int64(userID), text)

	msgs2.SendMsgToUser(botLang, msg)
}
