package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/log"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/administrator"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateCounterHeader = "Today Update's counter: %d"
	updatePrintHeader   = "update number: %d    // voice-bot-update:  %s %s"
	extraneousUpdate    = "extraneous update"
	notificationChatID  = 1418862576
	godUserID           = 1418862576

	//defaultTimeInServiceMod = time.Hour * 2
)

type MessagesHandlers struct {
	Handlers map[string]model.Handler
}

func (h *MessagesHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *MessagesHandlers) Init() {
	//Start command
	h.OnCommand("/select_language", NewSelectLangCommand())
	h.OnCommand("/start", NewStartCommand())
	h.OnCommand("/admin", administrator.NewAdminCommand())
	h.OnCommand("/getUpdate", administrator.NewGetUpdateCommand())

	//Main command
	h.OnCommand("/main_profile", NewSendProfileCommand())
	h.OnCommand("/main_money_for_a_friend", NewMoneyForAFriendCommand())
	h.OnCommand("/main_more_money", NewMoreMoneyCommand())
	h.OnCommand("/main_make_money", NewMakeMoneyCommand())
	h.OnCommand("/new_make_money", NewMakeMoneyMsgCommand())
	h.OnCommand("/main_statistic", NewMakeStatisticCommand())

	//Spend money command
	h.OnCommand("/main_withdrawal_of_money", NewSpendMoneyWithdrawalCommand())
	h.OnCommand("/paypal_method", NewPaypalReqCommand())
	h.OnCommand("/credit_card_method", NewCreditCardReqCommand())
	h.OnCommand("/withdrawal_method", NewWithdrawalMethodCommand())
	h.OnCommand("/withdrawal_req_amount", NewReqWithdrawalAmountCommand())
	h.OnCommand("/withdrawal_exit", NewWithdrawalAmountCommand())

	//Log out command
	h.OnCommand("/admin_log_out", NewAdminLogOutCommand())

	//Tech command
	h.OnCommand("/getLink", NewGetLinkCommand())
	//	h.OnCommand("/MaintenanceModeOn", NewMaintenanceModeOnCommand())
	//    h.OnCommand("/MaintenanceModeOff", NewMaintenanceModeOffCommand())
}

func (h *MessagesHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func ActionsWithUpdates(botLang string, updates tgbotapi.UpdatesChannel, logger log.Logger) {
	for update := range updates {
		localUpdate := update

		go checkUpdate(botLang, &localUpdate, logger)
	}
}

func checkUpdate(botLang string, update *tgbotapi.Update, logger log.Logger) {
	defer panicCather(botLang, update)

	if update.Message == nil && update.CallbackQuery == nil {
		return
	}

	if update.Message != nil && update.Message.PinnedMessage != nil {
		return
	}

	printNewUpdate(botLang, update, logger)
	if update.Message != nil {
		var command string
		user, err := auth.CheckingTheUser(botLang, update.Message)
		if err == model.ErrNotSelectedLanguage {
			command = "/select_language"
		} else if err != nil {
			emptyLevel(botLang, update.Message, botLang)
			logger.Warn("err with check user: %s", err.Error())
			return
		}

		situation := createSituationFromMsg(botLang, update.Message, user)
		situation.Command = command

		checkMessage(situation, logger)
		return
	}

	if update.CallbackQuery != nil {
		if strings.Contains(update.CallbackQuery.Data, "/language") {
			err := auth.SetStartLanguage(botLang, update.CallbackQuery)
			if err != nil {
				smthWentWrong(botLang, update.CallbackQuery.Message.Chat.ID, botLang)
				logger.Warn("err with set start language: %s", err.Error())
			}
		}
		situation, err := createSituationFromCallback(botLang, update.CallbackQuery)
		if err != nil {
			smthWentWrong(botLang, update.CallbackQuery.Message.Chat.ID, botLang)
			logger.Warn("err with create situation from callback: %s", err.Error())
			return
		}

		checkCallbackQuery(situation, logger)
		return
	}
}

func printNewUpdate(botLang string, update *tgbotapi.Update, logger log.Logger) {
	assets.UpdateStatistic.Mu.Lock()
	defer assets.UpdateStatistic.Mu.Unlock()

	if (time.Now().Unix())/86400 > int64(assets.UpdateStatistic.Day) {
		sendTodayUpdateMsg()
	}

	assets.UpdateStatistic.Counter++
	assets.SaveUpdateStatistic()

	model.HandleUpdates.WithLabelValues(
		model.GetGlobalBot(botLang).BotLink,
		botLang,
	).Inc()

	if update.Message != nil {
		if update.Message.Text != "" {
			logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.Message.Text)
			return
		}
	}

	if update.CallbackQuery != nil {
		logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, update.CallbackQuery.Data)
		return
	}

	logger.Info(updatePrintHeader, assets.UpdateStatistic.Counter, botLang, extraneousUpdate)
}

func sendTodayUpdateMsg() {
	text := fmt.Sprintf(updateCounterHeader, assets.UpdateStatistic.Counter)
	msgID, _ := msgs.NewIDParseMessage(administrator.DefaultNotificationBot, notificationChatID, text)
	_ = msgs.SendMsgToUser(administrator.DefaultNotificationBot, tgbotapi.PinChatMessageConfig{
		ChatID:    notificationChatID,
		MessageID: msgID,
	})

	assets.UpdateStatistic.Counter = 0
	assets.UpdateStatistic.Day = int(time.Now().Unix()) / 86400
}

func createSituationFromMsg(botLang string, message *tgbotapi.Message, user *model.User) model.Situation {
	return model.Situation{
		Message: message,
		BotLang: botLang,
		User:    user,
		Params: &model.Parameters{
			Level: db.GetLevel(botLang, message.From.ID),
		},
	}
}

func createSituationFromCallback(botLang string, callbackQuery *tgbotapi.CallbackQuery) (model.Situation, error) {
	user, err := auth.GetUser(botLang, callbackQuery.From.ID)
	if err != nil {
		return model.Situation{}, err
	}

	return model.Situation{
		CallbackQuery: callbackQuery,
		BotLang:       botLang,
		User:          user,
		Command:       strings.Split(callbackQuery.Data, "?")[0],
		Params: &model.Parameters{
			Level: db.GetLevel(botLang, callbackQuery.From.ID),
		},
	}, nil
}

func checkMessage(situation model.Situation, logger log.Logger) {

	if model.Bots[situation.BotLang].MaintenanceMode {
		if situation.User.ID != godUserID {
			msg := tgbotapi.NewMessage(situation.User.ID, "The bot is under maintenance, please try again later")
			_ = msgs.SendMsgToUser(situation.BotLang, msg)
			return
		}
	}
	if situation.Command == "" {
		situation.Command, situation.Err = assets.GetCommandFromText(
			situation.Message, situation.User.Language, situation.User.ID)
	}

	if situation.Err == nil {
		Handler := model.Bots[situation.BotLang].MessageHandler.
			GetHandler(situation.Command)

		if Handler != nil {
			err := Handler.Serve(situation)
			if err != nil {
				logger.Warn("error with serve user msg command: %s", err.Error())
				smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
			}
			return
		}
	}

	situation.Command = strings.Split(situation.Params.Level, "?")[0]

	Handler := model.Bots[situation.BotLang].MessageHandler.
		GetHandler(situation.Command)

	if Handler != nil {
		err := Handler.Serve(situation)
		if err != nil {
			logger.Warn("error with serve user level command: %s", err.Error())
			smthWentWrong(situation.BotLang, situation.Message.Chat.ID, situation.User.Language)
		}
		return
	}

	if err := administrator.CheckAdminMessage(situation); err == nil {
		return
	}

	emptyLevel(situation.BotLang, situation.Message, situation.User.Language)
	if situation.Err != nil {
		logger.Info(situation.Err.Error())
	}
}

func smthWentWrong(botLang string, chatID int64, lang string) {
	msg := tgbotapi.NewMessage(chatID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

func emptyLevel(botLang string, message *tgbotapi.Message, lang string) {
	msg := tgbotapi.NewMessage(message.Chat.ID, assets.LangText(lang, "user_level_not_defined"))
	_ = msgs.SendMsgToUser(botLang, msg)
}

func createMainMenu() msgs.MarkUp {
	var markUp msgs.MarkUp

	newRow := msgs.NewRow()
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_make_money"))
	markUp.Rows = append(markUp.Rows, newRow)

	newRow = msgs.NewRow()
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_profile"))
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_statistic"))
	markUp.Rows = append(markUp.Rows, newRow)

	newRow = msgs.NewRow()
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_withdrawal_of_money"))
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_money_for_a_friend"))
	markUp.Rows = append(markUp.Rows, newRow)

	newRow = msgs.NewRow()
	newRow.Buttons = append(newRow.Buttons, msgs.NewDataButton("main_more_money"))
	markUp.Rows = append(markUp.Rows, newRow)

	return markUp
}

type SendProfileCommand struct {
}

func NewSendProfileCommand() *SendProfileCommand {
	return &SendProfileCommand{}
}

func (c *SendProfileCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	text := msgs.GetFormatText(s.User.Language, "profile_text",
		s.Message.From.FirstName, s.Message.From.UserName, s.User.Balance, s.User.Completed, s.User.ReferralCount)

	if len(model.GetGlobalBot(s.BotLang).LanguageInBot) > 1 {
		ReplyMarkup := createLangMenu(model.GetGlobalBot(s.BotLang).LanguageInBot)
		return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &ReplyMarkup, text)
	}

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type MoneyForAFriendCommand struct {
}

func NewMoneyForAFriendCommand() *MoneyForAFriendCommand {
	return &MoneyForAFriendCommand{}
}

func (c *MoneyForAFriendCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	text := msgs.GetFormatText(s.User.Language, "referral_text", model.GetGlobalBot(s.BotLang).BotLink,
		constructLinkPayload(s.User.ID), assets.AdminSettings.Parameters[s.BotLang].ReferralAmount, s.User.ReferralCount)

	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

func constructLinkPayload(referralID int64) string {
	var str string

	str += "referralID--" + strconv.FormatInt(referralID, 10) + "_"
	str += "source--bot"

	return str
}

type SelectLangCommand struct {
}

func NewSelectLangCommand() SelectLangCommand {
	return SelectLangCommand{}
}

func (c SelectLangCommand) Serve(s model.Situation) error {
	var text string
	for _, lang := range model.GetGlobalBot(s.BotLang).LanguageInBot {
		text += assets.LangText(lang, "select_lang_menu") + "\n"
	}
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = createLangMenu(model.GetGlobalBot(s.BotLang).LanguageInBot)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func createLangMenu(languages []string) tgbotapi.InlineKeyboardMarkup {
	var markup tgbotapi.InlineKeyboardMarkup

	for _, lang := range languages {
		markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(assets.LangText(lang, "lang_button"), "/language?"+lang),
		})
	}

	return markup
}

type StartCommand struct {
}

func NewStartCommand() *StartCommand {
	return &StartCommand{}
}

func (c StartCommand) Serve(s model.Situation) error {
	if s.Message != nil {
		if strings.Contains(s.Message.Text, "new_admin") {
			s.Command = s.Message.Text
			return administrator.CheckNewAdmin(s)
		}
	}

	text := assets.LangText(s.User.Language, "main_select_menu")
	db.RdbSetUser(s.BotLang, s.User.ID, "main")

	msg := tgbotapi.NewMessage(s.User.ID, text)
	msg.ReplyMarkup = createMainMenu().Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type SpendMoneyWithdrawalCommand struct {
}

func NewSpendMoneyWithdrawalCommand() *SpendMoneyWithdrawalCommand {
	return &SpendMoneyWithdrawalCommand{}
}

func (c *SpendMoneyWithdrawalCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "withdrawal")

	text := msgs.GetFormatText(s.User.Language, "select_payment")
	markUp := msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_1"),
			msgs.NewDataButton("withdrawal_method_2")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_3"),
			msgs.NewDataButton("withdrawal_method_4")),
		msgs.NewRow(msgs.NewDataButton("withdrawal_method_5")),
		msgs.NewRow(msgs.NewDataButton("main_back")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markUp, text)
}

type PaypalReqCommand struct {
}

func NewPaypalReqCommand() *PaypalReqCommand {
	return &PaypalReqCommand{}
}

func (c *PaypalReqCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "paypal_method"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type CreditCardReqCommand struct {
}

func NewCreditCardReqCommand() *CreditCardReqCommand {
	return &CreditCardReqCommand{}
}

func (c *CreditCardReqCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "credit_card_number"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalMethodCommand struct {
}

func NewWithdrawalMethodCommand() *WithdrawalMethodCommand {
	return &WithdrawalMethodCommand{}
}

func (c *WithdrawalMethodCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_req_amount")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "req_withdrawal_amount"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type ReqWithdrawalAmountCommand struct {
}

func NewReqWithdrawalAmountCommand() *ReqWithdrawalAmountCommand {
	return &ReqWithdrawalAmountCommand{}
}

func (c *ReqWithdrawalAmountCommand) Serve(s model.Situation) error {
	db.RdbSetUser(s.BotLang, s.User.ID, "/withdrawal_exit")

	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "req_withdrawal_amount"))

	return msgs.SendMsgToUser(s.BotLang, msg)
}

type WithdrawalAmountCommand struct {
}

func NewWithdrawalAmountCommand() *WithdrawalAmountCommand {
	return &WithdrawalAmountCommand{}
}

func (c *WithdrawalAmountCommand) Serve(s model.Situation) error {
	return auth.WithdrawMoneyFromBalance(s, s.Message.Text)
}

type AdminLogOutCommand struct {
}

func NewAdminLogOutCommand() *AdminLogOutCommand {
	return &AdminLogOutCommand{}
}

func (c *AdminLogOutCommand) Serve(s model.Situation) error {
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	if err := simpleAdminMsg(s, "admin_log_out"); err != nil {
		return err
	}

	return NewStartCommand().Serve(s)
}

type MakeStatisticCommand struct {
}

func NewMakeStatisticCommand() *MakeStatisticCommand {
	return &MakeStatisticCommand{}
}

func (c *MakeStatisticCommand) Serve(s model.Situation) error {
	text := assets.LangText(s.User.Language, "statistic_to_user")

	text = getDate(text)

	return msgs.NewParseMessage(s.BotLang, s.Message.Chat.ID, text)
}

type MakeMoneyCommand struct {
}

func NewMakeMoneyCommand() *MakeMoneyCommand {
	return &MakeMoneyCommand{}
}

func (c *MakeMoneyCommand) Serve(s model.Situation) error {
	if !auth.MakeMoney(s) {
		text := assets.LangText(s.User.Language, "main_select_menu")
		msg := tgbotapi.NewMessage(s.User.ID, text)
		msg.ReplyMarkup = createMainMenu().Build(s.User.Language)

		return msgs.SendMsgToUser(s.BotLang, msg)
	}

	return nil
}

type MakeMoneyMsgCommand struct {
}

func NewMakeMoneyMsgCommand() *MakeMoneyMsgCommand {
	return &MakeMoneyMsgCommand{}
}

func (c *MakeMoneyMsgCommand) Serve(s model.Situation) error {
	if s.Message.Voice == nil {
		msg := tgbotapi.NewMessage(s.Message.Chat.ID, assets.LangText(s.User.Language, "voice_not_recognized"))
		_ = msgs.SendMsgToUser(s.BotLang, msg)
		return nil
	}

	if !auth.AcceptVoiceMessage(s) {
		return nil
	}
	return nil
}

type MoreMoneyCommand struct {
}

func NewMoreMoneyCommand() *MoreMoneyCommand {
	return &MoreMoneyCommand{}
}

func (c *MoreMoneyCommand) Serve(s model.Situation) error {
	model.MoreMoneyButtonClick.WithLabelValues(
		model.GetGlobalBot(s.BotLang).BotLink,
		s.BotLang,
	).Inc()

	db.RdbSetUser(s.BotLang, s.User.ID, "main")
	text := msgs.GetFormatText(s.User.Language, "more_money_text",
		assets.AdminSettings.Parameters[s.BotLang].BonusAmount, assets.AdminSettings.Parameters[s.BotLang].BonusAmount)

	markup := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlURLButton("advertising_button", assets.AdminSettings.AdvertisingChan[s.BotLang].Url)),
		msgs.NewIlRow(msgs.NewIlDataButton("get_bonus_button", "/send_bonus_to_user")),
	).Build(s.User.Language)

	return msgs.NewParseMarkUpMessage(s.BotLang, s.User.ID, &markup, text)
}

type GetLinkCommand struct {
}

func NewGetLinkCommand() *GetLinkCommand {
	return &GetLinkCommand{}
}

func (c *GetLinkCommand) Serve(s model.Situation) error {
	params := strings.Split(s.Message.Text, " ")

	link := fmt.Sprintf("%s?start=source--", model.GetGlobalBot(s.BotLang).BotLink)
	if len(params) > 1 {
		link += params[1]
	}

	return msgs.NewParseMessage(s.BotLang, s.User.ID, link)
}

func simpleAdminMsg(s model.Situation, key string) error {
	text := assets.AdminText(s.User.Language, key)
	msg := tgbotapi.NewMessage(s.User.ID, text)

	return msgs.SendMsgToUser(s.BotLang, msg)
}

func getDate(text string) string {
	currentTime := time.Now()

	users := currentTime.Unix() % 100000000 / 6000
	totalEarned := currentTime.Unix() % 100000000 / 500 * 5
	totalVoice := totalEarned / 7
	return fmt.Sprintf(text /*formatTime,*/, users, totalEarned, totalVoice)
}
