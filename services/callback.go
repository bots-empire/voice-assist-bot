package services

import (
	"strconv"
	"strings"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/log"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"github.com/Stepan1328/voice-assist-bot/services/administrator"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallBackHandlers struct {
	Handlers map[string]model.Handler
}

func (h *CallBackHandlers) GetHandler(command string) model.Handler {
	return h.Handlers[command]
}

func (h *CallBackHandlers) Init() {
	//Money command
	h.OnCommand("/language", NewLanguageCommand())
	h.OnCommand("/send_bonus_to_user", NewGetBonusCommand())
	h.OnCommand("/withdrawal_money", NewRecheckSubscribeCommand())
	h.OnCommand("/promotion_case", NewPromotionCaseCommand())
}

func (h *CallBackHandlers) OnCommand(command string, handler model.Handler) {
	h.Handlers[command] = handler
}

func checkCallbackQuery(s model.Situation, logger log.Logger) {
	if strings.Contains(s.Params.Level, "admin") {
		if err := administrator.CheckAdminCallback(s); err != nil {
			logger.Warn("error with serve admin callback command: %s", err.Error())
		}
		return
	}

	Handler := model.Bots[s.BotLang].CallbackHandler.
		GetHandler(s.Command)

	if Handler != nil {
		if err := Handler.Serve(s); err != nil {
			logger.Warn("error with serve user callback command: %s", err.Error())
			smthWentWrong(s.BotLang, s.CallbackQuery.Message.Chat.ID, s.User.Language)
		}
		return
	}

	logger.Warn("get callback data='%s', but they didn't react in any way", s.CallbackQuery.Data)
}

type LanguageCommand struct {
}

func NewLanguageCommand() *LanguageCommand {
	return &LanguageCommand{}
}

func (c *LanguageCommand) Serve(s model.Situation) error {
	lang := strings.Split(s.CallbackQuery.Data, "?")[1]

	level := db.GetLevel(s.BotLang, s.User.ID)
	if strings.Contains(level, "admin") {
		return nil
	}

	s.User.Language = lang

	return NewStartCommand().Serve(s)
}

type GetBonusCommand struct {
}

func NewGetBonusCommand() *GetBonusCommand {
	return &GetBonusCommand{}
}

func (c *GetBonusCommand) Serve(s model.Situation) error {
	return auth.GetABonus(s)
}

type RecheckSubscribeCommand struct {
}

func NewRecheckSubscribeCommand() *RecheckSubscribeCommand {
	return &RecheckSubscribeCommand{}
}

func (c *RecheckSubscribeCommand) Serve(s model.Situation) error {
	amount := strings.Split(s.CallbackQuery.Data, "?")[1]
	s.Message = &tgbotapi.Message{
		Text: amount,
	}
	if err := msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_subscribe"); err != nil {
		return err
	}
	amountInt, _ := strconv.Atoi(amount)

	if auth.CheckSubscribeToWithdrawal(s, amountInt) {
		db.RdbSetUser(s.BotLang, s.User.ID, "main")

		return NewStartCommand().Serve(s)
	}
	return nil
}

type PromotionCaseCommand struct {
}

func NewPromotionCaseCommand() *PromotionCaseCommand {
	return &PromotionCaseCommand{}
}

func (c *PromotionCaseCommand) Serve(s model.Situation) error {
	cost, err := strconv.Atoi(strings.Split(s.CallbackQuery.Data, "?")[1])
	if err != nil {
		return err
	}

	if s.User.Balance < cost {
		return msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "not_enough_money")
	}

	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)
	msg := tgbotapi.NewMessage(s.User.ID, assets.LangText(s.User.Language, "invitation_to_send_link_text"))
	msg.ReplyMarkup = msgs.NewMarkUp(
		msgs.NewRow(msgs.NewDataButton("withdraw_cancel")),
	).Build(s.User.Language)

	if err := msgs.SendAnswerCallback(s.BotLang, s.CallbackQuery, s.User.Language, "invitation_to_send_link"); err != nil {
		return err
	}

	return msgs.SendMsgToUser(s.BotLang, msg)
}
