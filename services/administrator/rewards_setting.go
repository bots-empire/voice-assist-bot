package administrator

import (
	"github.com/bots-empire/base-bot/msgs"
	"strings"

	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	bonusAmount         = "bonus_amount"
	minWithdrawalAmount = "min_withdrawal_amount"
	voiceAmount         = "voice_amount"
	voicePDAmount       = "voice_pd_amount"
	referralAmount      = "referral_amount"
	currencyType        = "currency_type"
)

func (a *Admin) MakeMoneySettingCommand(s *model.Situation) error {
	markUp, text := a.sendMakeMoneyMenu(s.BotLang, s.User.ID)

	if db.RdbGetAdminMsgID(s.BotLang, s.User.ID) != 0 {
		err := a.msgs.NewEditMarkUpMessage(s.User.ID, db.RdbGetAdminMsgID(s.BotLang, s.User.ID), markUp, text)
		if err != nil {
			return errors.Wrap(err, "failed to edit markup")
		}
		err = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "make_a_choice")
		if err != nil {
			return errors.Wrap(err, "failed to send admin answer callback")
		}
		return nil
	}
	msgID, err := a.msgs.NewIDParseMarkUpMessage(s.User.ID, markUp, text)
	if err != nil {
		return errors.Wrap(err, "failed parse new id markup message")
	}
	db.RdbSetAdminMsgID(s.BotLang, s.User.ID, msgID)
	return nil
}

func (a *Admin) sendMakeMoneyMenu(botLang string, userID int64) (*tgbotapi.InlineKeyboardMarkup, string) {
	lang := model.AdminLang(userID)
	text := a.bot.AdminText(lang, "make_money_setting_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("change_bonus_amount_button", "admin/make_money?bonus_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_min_withdrawal_amount_button", "admin/make_money?min_withdrawal_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_voice_amount_button", "admin/make_money?voice_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_voice_pd_amount_button", "admin/make_money?voice_pd_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_referral_amount_button", "admin/make_money?referral_amount")),
		msgs.NewIlRow(msgs.NewIlAdminButton("change_currency_type_button", "admin/make_money?currency_type")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_main_menu", "admin/send_menu")),
	).Build(a.bot.AdminLibrary[lang])

	db.RdbSetUser(botLang, userID, "admin/make_money_settings")
	return &markUp, text
}

func (a *Admin) ChangeParameterCommand(s *model.Situation) error {
	changeParameter := strings.Split(s.CallbackQuery.Data, "?")[1]

	lang := model.AdminLang(s.User.ID)
	var parameter, text string
	var value interface{}

	db.RdbSetUser(s.BotLang, s.User.ID, "admin/make_money?"+changeParameter)

	switch changeParameter {
	case bonusAmount:
		parameter = a.bot.AdminText(lang, "change_bonus_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).BonusAmount
	case minWithdrawalAmount:
		parameter = a.bot.AdminText(lang, "change_min_withdrawal_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).MinWithdrawalAmount
	case voiceAmount:
		parameter = a.bot.AdminText(lang, "change_voice_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).VoiceAmount
	case voicePDAmount:
		parameter = a.bot.AdminText(lang, "change_voice_pd_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).MaxOfVoicePerDay
	case referralAmount:
		parameter = a.bot.AdminText(lang, "change_referral_amount_button")
		value = model.AdminSettings.GetParams(s.BotLang).ReferralAmount
	case currencyType:
		parameter = a.bot.AdminText(lang, "change_currency_type_button")
		value = model.AdminSettings.GetCurrency(s.BotLang)
	}

	text = a.adminFormatText(lang, "set_new_amount_text", parameter, value)
	_ = a.msgs.SendAdminAnswerCallback(s.CallbackQuery, "type_the_text")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewAdminButton("back_to_make_money_setting")),
		msgs.NewRow(msgs.NewAdminButton("exit")),
	).Build(a.bot.AdminLibrary[lang])

	return a.msgs.NewParseMarkUpMessage(s.User.ID, markUp, text)
}
