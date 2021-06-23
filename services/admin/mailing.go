package admin

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

func analyzeMailingCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "advertisement/", "", 1)
	if strings.Contains(callbackQuery.Data, "/") {
		analyzeSelectedMailingCallbackLevel(botLang, callbackQuery)
		return
	}

	db.RdbSetUser(botLang, callbackQuery.From.ID, "admin/advertisement/change_text")
	resetSelectedLang()
	sendMailingMenu(botLang, callbackQuery.From.ID)
	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
}

func analyzeSelectedMailingCallbackLevel(botLang string, callbackQuery *tgbotapi.CallbackQuery) {
	callbackQuery.Data = strings.Replace(callbackQuery.Data, "mailing/", "", 1)
	switch callbackQuery.Data {
	case "send":
		//if !selectedLangAreNotEmpty() {
		//	msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "no_language_selected")
		//	return
		//}
		go db.StartMailing(botLang)
		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "mailing_successful")
		resendAdvertisementMenuLevel(botLang, callbackQuery.From.ID)
	case "back":
		callbackQuery.Data = "advertisement"
		settingAdvertisementCallbackLevel(botLang, callbackQuery)
	case "select_all", "deselect_all":
		switchedSelectedLanguages()
		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		sendMailingMenu(botLang, callbackQuery.From.ID)
	default:
		switchLangOnKeyboard(callbackQuery.Data)
		msgs2.SendAdminAnswerCallback(botLang, callbackQuery, "make_a_choice")
		sendMailingMenu(botLang, callbackQuery.From.ID)
	}
}

func sendMailingMenu(botLang string, userID int) {
	lang := assets.AdminLang(userID)

	text := assets.AdminText(lang, "mailing_main_text")
	markUp := createMailingMarkUp(lang)

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID := msgs2.NewIDParseMarkUpMessage(botLang, int64(userID), &markUp, text)
		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return
	}
	msgs2.NewEditMarkUpMessage(botLang, userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
}

func createMailingMarkUp(lang string) tgbotapi.InlineKeyboardMarkup {
	//markUp := parseMainLanguageButton("mailing")
	//
	//text := "select_all_language"
	//data := "admin/advertisement/mailing/select_all"
	//if selectedAllLanguage() {
	//	text = "deselect_all_selections"
	//	data = strings.Replace(data, "select_all", "deselect_all", 1)
	//}
	//
	//markUp.Rows = append(markUp.Rows,
	//	msgs2.NewIlRow(msgs2.NewIlAdminButton(text, data)),
	//	msgs2.NewIlRow(msgs2.NewIlAdminButton("start_mailing_button", "admin/advertisement/mailing/send")),
	//	msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement/mailing/back")),
	//)

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("start_mailing_button", "admin/advertisement/mailing/send")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_advertisement_setting", "admin/advertisement/mailing/back")),
	)

	return markUp.Build(lang)
}

func switchLangOnKeyboard(lang string) {
	assets.AdminSettings.LangSelectedMap[lang] = !assets.AdminSettings.LangSelectedMap[lang]
	assets.SaveAdminSettings()
}

func switchedSelectedLanguages() {
	if selectedAllLanguage() {
		resetSelectedLang()
		return
	}
	chooseAllLanguages()
}

func resetSelectedLang() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = false
	}
	assets.SaveAdminSettings()
}

func chooseAllLanguages() {
	for lang := range assets.AdminSettings.LangSelectedMap {
		assets.AdminSettings.LangSelectedMap[lang] = true
	}
	assets.SaveAdminSettings()
}

func selectedAllLanguage() bool {
	for _, lang := range assets.AvailableLang {
		if !assets.AdminSettings.LangSelectedMap[lang] {
			return false
		}
	}
	return true
}

func selectedLangAreNotEmpty() bool {
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			return true
		}
	}
	return false
}
