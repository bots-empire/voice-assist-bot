package administrator

import (
	"strconv"
	"strings"

	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminMessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *AdminMessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *AdminMessagesHandlers) Init(adminSrv *Admin) {
	//Delete Admin command
	h.OnCommand("/delete_admin", adminSrv.RemoveAdminCommand)

	//Change Advertisement parameters command
	h.OnCommand("/make_money", adminSrv.UpdateParameterCommand)
	h.OnCommand("/change_text_url", adminSrv.SetNewTextUrlCommand)
	h.OnCommand("/advertisement_setting", adminSrv.AdvertisementSettingCommand)
	h.OnCommand("/get_new_source", adminSrv.GetNewSourceCommand)
}

func (h *AdminMessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func (a *Admin) RemoveAdminCommand(s *model.Situation) error {
	lang := model.AdminLang(s.User.ID)
	adminId, err := strconv.ParseInt(s.Message.Text, 10, 64)
	if err != nil {
		text := a.bot.AdminText(lang, "incorrect_admin_id_text")
		return a.msgs.NewParseMessage(s.User.ID, text)
	}

	if !checkAdminIDInTheList(adminId) {
		text := a.bot.AdminText(lang, "incorrect_admin_id_text")
		return a.msgs.NewParseMessage(s.User.ID, text)

	}

	delete(model.AdminSettings.AdminID, adminId)
	model.SaveAdminSettings()
	if err := a.setAdminBackButton(s.User.ID, "admin_removed_status"); err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.Command = "admin/send_admin_list"
	s.CallbackQuery = &tgbotapi.CallbackQuery{Data: "admin/send_admin_list"}
	return a.AdminListCommand(s)
}

func checkAdminIDInTheList(adminID int64) bool {
	_, inMap := model.AdminSettings.AdminID[adminID]
	return inMap
}

func (a *Admin) UpdateParameterCommand(s *model.Situation) error {
	if strings.Contains(s.Params.Level, "make_money?") && s.Message.Text == "← Назад к ⚙️ Заработок" {
		if err := a.setAdminBackButton(s.User.ID, "operation_canceled"); err != nil {
			return err
		}
		db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
		s.Command = "admin/make_money_setting"

		return a.MakeMoneySettingCommand(s)
	}

	partition := strings.Split(s.Params.Level, "?")[1]

	if partition == currencyType {
		model.AdminSettings.UpdateCurrency(s.BotLang, s.Message.Text)
	} else {
		err := a.setNewIntParameter(s, partition)
		if err != nil {
			return err
		}
	}

	model.SaveAdminSettings()
	err := a.setAdminBackButton(s.User.ID, "operation_completed")
	if err != nil {
		return nil
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	s.Command = "admin/make_money_setting"

	return a.MakeMoneySettingCommand(s)
}

func (a *Admin) setNewIntParameter(s *model.Situation, partition string) error {
	lang := model.AdminLang(s.User.ID)

	newAmount, err := strconv.Atoi(s.Message.Text)
	if err != nil || newAmount <= 0 {
		text := a.bot.AdminText(lang, "incorrect_make_money_change_input")
		return a.msgs.NewParseMessage(s.User.ID, text)
	}

	switch partition {
	case bonusAmount:
		model.AdminSettings.UpdateBonusAmount(s.BotLang, newAmount)
	case minWithdrawalAmount:
		model.AdminSettings.UpdateMinWithdrawalAmount(s.BotLang, newAmount)
	case voiceAmount:
		model.AdminSettings.UpdateVoiceAmount(s.BotLang, newAmount)
	case voicePDAmount:
		model.AdminSettings.UpdateMaxOfVoicePerDay(s.BotLang, newAmount)
	case referralAmount:
		model.AdminSettings.UpdateReferralAmount(s.BotLang, newAmount)
	}

	return nil
}

func (a *Admin) SetNewTextUrlCommand(s *model.Situation) error {
	capitation := strings.Split(s.Params.Level, "?")[1]
	channel, _ := strconv.Atoi(strings.Split(s.Params.Level, "?")[2])
	lang := model.AdminLang(s.User.ID)
	status := "operation_canceled"

	switch capitation {
	case "change_url":
		url, chatID := getUrlAndChatID(s.Message)
		if chatID == 0 {
			text := a.bot.AdminText(lang, "chat_id_not_update")
			return a.msgs.NewParseMessage(s.User.ID, text)
		}
		model.AdminSettings.UpdateAdvertChannelID(s.BotLang, chatID, channel)
		model.AdminSettings.UpdateAdvertUrl(s.BotLang, channel, url)
		//assets.AdminSettings.UpdateAdvertChan(s.BotLang, advertChan)
	case "change_text":
		model.AdminSettings.UpdateAdvertText(s.BotLang, s.Message.Text, channel)
	case "change_photo":
		if len(s.Message.Photo) == 0 {
			text := a.bot.AdminText(lang, "send_only_photo")
			return a.msgs.NewParseMessage(s.User.ID, text)
		}
		model.AdminSettings.UpdateAdvertPhoto(s.BotLang, channel, s.Message.Photo[0].FileID)
	case "change_video":
		if s.Message.Video == nil {
			text := a.bot.AdminText(lang, "send_only_video")
			return a.msgs.NewParseMessage(s.User.ID, text)
		}
		model.AdminSettings.UpdateAdvertVideo(s.BotLang, channel, s.Message.Video.FileID)
	}

	model.SaveAdminSettings()
	status = "operation_completed"

	if err := a.setAdminBackButton(s.User.ID, status); err != nil {
		return err
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "admin")
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)

	s.Command = "admin/advertisement"
	s.Params.Level = "admin/change_url"
	return a.AdvertisementMenuCommand(s)
}

func (a *Admin) AdvertisementSettingCommand(s *model.Situation) error {
	s.CallbackQuery = &tgbotapi.CallbackQuery{
		Data: "admin/change_text_url?",
	}
	s.Command = "admin/advertisement"
	return a.AdvertisementMenuCommand(s)
}

func getUrlAndChatID(message *tgbotapi.Message) (string, int64) {
	data := strings.Split(message.Text, "\n")
	if len(data) != 2 {
		return "", 0
	}

	chatId, err := strconv.Atoi(data[0])
	if err != nil {
		return "", 0
	}

	//advert := &assets.AdvertChannel{
	//	Url:       map[int]string{channel: data[1]},
	//	ChannelID: int64(chatId),
	//}

	//advert.Url[channel] = data[1]
	//advert.ChannelID = int64(chatId)

	return data[1], int64(chatId)
}

func (a *Admin) CheckAdminMessage(s *model.Situation) error {
	if !ContainsInAdmin(s.User.ID) {
		return a.notAdmin(s.User)
	}

	s.Command, s.Err = a.bot.GetCommandFromText(s.Message, s.User.Language, s.User.ID)
	if s.Err == nil {
		Handler := model.Bots[s.BotLang].AdminMessageHandler.
			GetHandler(s.Command)

		if Handler != nil {
			return Handler(s)
		}
	}

	s.Command = strings.TrimLeft(strings.Split(s.Params.Level, "?")[0], "admin")

	Handler := model.Bots[s.BotLang].AdminMessageHandler.
		GetHandler(s.Command)

	if Handler != nil {
		return Handler(s)
	}

	return model.ErrCommandNotConverted
}
