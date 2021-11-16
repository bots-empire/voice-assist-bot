package administrator

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type AdminMessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminMessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *AdminMessagesHandlers) Init() {
	h.OnCommand("/make_money", NewUpdateParameterCommand())
	h.OnCommand("/change_text_url", NewSetNewTextUrlCommand())
	h.OnCommand("/advertisement_setting", NewAdvertisementSettingCommand())
}

func (h *AdminMessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

type UpdateParameterCommand struct {
}

func NewUpdateParameterCommand() *UpdateParameterCommand {
	return &UpdateParameterCommand{}
}

func (c *UpdateParameterCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	partition := strings.Split(s.Params.Level, "?")[1]
	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := assets.AdminText(lang, "incorrect_make_money_change_input")
		return msgs.NewParseMessage(s.BotLang, int64(s.User.ID), text)
	}
	switch partition {
	case bonusAmount:
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount = newAmount
	case minWithdrawalAmount:
		assets.AdminSettings.Parameters[s.BotLang].MinWithdrawalAmount = newAmount
	case voiceAmount:
		assets.AdminSettings.Parameters[s.BotLang].VoiceAmount = newAmount
	case voicePDAmount:
		assets.AdminSettings.Parameters[s.BotLang].MaxOfVoicePerDay = newAmount
	case referralAmount:
		assets.AdminSettings.Parameters[s.BotLang].ReferralAmount = newAmount
	}

	assets.SaveAdminSettings()
	err = setAdminBackButton(s.BotLang, s.User.ID, "operation_completed")
	if err != nil {
		return errors.Wrap(err, "cannot set admin back button")
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.Command = "admin/make_money_setting"
	err = CheckAdminCallback(s)
	if err != nil {
		return errors.Wrap(err, "cannot check admin callback")
	}

	return NewMakeMoneySettingCommand().Serve(s)
}

type SetNewTextUrlCommand struct {
}

func NewSetNewTextUrlCommand() *SetNewTextUrlCommand {
	return &SetNewTextUrlCommand{}
}

func (c *SetNewTextUrlCommand) Serve(s model.Situation) error {
	capitation := strings.Split(s.Params.Level, "?")[1]
	lang := assets.AdminLang(s.User.ID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		advertChan := getUrlAndChatID(s.Message)
		if advertChan.ChannelID == 0 {
			text := assets.AdminText(lang, "chat_id_not_update")
			return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
		}

		assets.AdminSettings.AdvertisingChan[s.BotLang] = advertChan
	case "change_text":
		assets.AdminSettings.AdvertisingText[s.BotLang] = s.Message.Text
	}
	assets.SaveAdminSettings()
	status = "operation_completed"

	if err := setAdminBackButton(s.BotLang, s.User.ID, status); err != nil {
		return err
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.Command = "admin/advertisement"
	s.Params.Level = "admin/change_url"
	return NewAdvertisementMenuCommand().Serve(s)
}

type AdvertisementSettingCommand struct {
}

func NewAdvertisementSettingCommand() *AdvertisementSettingCommand {
	return &AdvertisementSettingCommand{}
}

func (c *AdvertisementSettingCommand) Serve(s model.Situation) error {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}
	s.Command = "admin/advertisement"
	return NewAdvertisementMenuCommand().Serve(s)
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

func CheckAdminMessage(s model.Situation) error {
	if !containsInAdmin(s.User.ID) {
		return notAdmin(s.BotLang, s.User)
	}

	s.Command, s.Err = assets.GetCommandFromText(s.Message, s.User.Language, s.User.ID)
	if s.Err == nil {
		Handler := model.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			return Handler.Serve(s)
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := model.Bots[s.BotLang].AdminMessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler.Serve(s)
	}

	return model.ErrCommandNotConverted
}
