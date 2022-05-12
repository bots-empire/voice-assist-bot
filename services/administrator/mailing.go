package administrator

import (
	"github.com/bots-empire/base-bot/msgs"
	"strconv"
	"strings"

	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (a *Admin) StartMailingCommand(s *model.Situation) error {
	channel, _ := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	go a.mailing.StartMailing(s.BotLang, s.User.ID, channel)

	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "mailing_successful")
	if channel == model.GlobalMailing {
		return a.AdvertisementMenuCommand(s)
	}
	return a.resendAdvertisementMenuLevel(s.BotLang, s.User.ID, channel)
}

func (a *Admin) SelectedLangCommand(s *model.Situation) error {
	data := strings.Split(s.CallbackQuery.Data, "?")
	partition := data[1]
	lang := data[2]
	switch partition {
	case "switch_lang":
		switchLangOnKeyboard(lang)
		if err := a.sendMailingMenu(s.BotLang, s.User.ID, "1"); err != nil {
			return err
		}
		return a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
	case "switch_all":
		a.switchedSelectedLanguages()
		if err := a.sendMailingMenu(s.BotLang, s.User.ID, "1"); err != nil {
			return err
		}

		return a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
	}
	return nil
}

func (a *Admin) sendMailingMenu(botLang string, userID int64, channel string) error {
	lang := model.AdminLang(userID)

	text := a.bot.AdminText(lang, "mailing_main_text")
	markUp := createMailingMarkUp(botLang, channel, a.bot.AdminLibrary[lang])

	if db.RdbGetAdminMsgID(botLang, userID) == 0 {
		msgID, err := a.msgs.NewIDParseMarkUpMessage(userID, &markUp, text)
		if err != nil {
			return err
		}

		db.RdbSetAdminMsgID(botLang, userID, msgID)
		return nil
	}

	return a.msgs.NewEditMarkUpMessage(userID, db.RdbGetAdminMsgID(botLang, userID), &markUp, text)
}

func createMailingMarkUp(botLang, channel string, texts map[string]string) tgbotapi.InlineKeyboardMarkup {
	markUp := &msgs.InlineMarkUp{}

	if buttonUnderAdvertisementUnable(botLang) {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("advert_button_on", "admin/change_advert_button_status?"+channel)),
		)
	} else {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("advert_button_off", "admin/change_advert_button_status?"+channel)),
		)
	}

	if channel == "4" {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("start_mailing_button", "admin/start_mailing?"+channel)),
			msgs.NewIlRow(msgs.NewIlAdminButton("back_to_chan_menu", "admin/advertisement")),
		)
	} else {
		markUp.Rows = append(markUp.Rows,
			msgs.NewIlRow(msgs.NewIlAdminButton("start_mailing_button", "admin/start_mailing?"+channel)),
			msgs.NewIlRow(msgs.NewIlAdminButton("back_to_advertisement_setting", "admin/change_advert_chan?"+channel)),
		)
	}

	return markUp.Build(texts)
}

func switchLangOnKeyboard(lang string) {
	model.AdminSettings.GlobalParameters[lang].LangSelectedMap[lang] = !model.AdminSettings.GlobalParameters[lang].LangSelectedMap[lang]
	model.SaveAdminSettings()
}

func (a *Admin) resendAdvertisementMenuLevel(botLang string, userID int64, channel int) error {
	db.DeleteOldAdminMsg(botLang, userID)

	db.RdbSetUser(botLang, userID, "admin/advertisement")
	inlineMarkUp, text := a.getAdvertisementMenu(botLang, userID, channel)
	msgID, err := a.msgs.NewIDParseMarkUpMessage(userID, inlineMarkUp, text)
	if err != nil {
		return err
	}
	db.RdbSetAdminMsgID(botLang, userID, msgID)
	return nil
}

func (a *Admin) switchedSelectedLanguages() {
	if a.selectedAllLanguage() {
		resetSelectedLang()
		return
	}
	chooseAllLanguages()
}

func resetSelectedLang() {
	for lang := range model.AdminSettings.GlobalParameters {
		model.AdminSettings.GlobalParameters[lang].LangSelectedMap[lang] = false
	}
	model.SaveAdminSettings()
}

func chooseAllLanguages() {
	for lang := range model.AdminSettings.GlobalParameters {
		model.AdminSettings.GlobalParameters[lang].LangSelectedMap[lang] = true
	}
	model.SaveAdminSettings()
}

func (a *Admin) selectedAllLanguage() bool {
	for _, lang := range a.bot.LanguageInBot {
		if !model.AdminSettings.GlobalParameters[lang].LangSelectedMap[lang] {
			return false
		}
	}
	return true
}

func buttonUnderAdvertisementUnable(botLang string) bool {
	return model.AdminSettings.GlobalParameters[botLang].Parameters.ButtonUnderAdvert
}
