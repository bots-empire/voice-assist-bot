package admin

import (
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

type Situation struct {
	Message       *tgbotapi.Message
	CallbackQuery *tgbotapi.CallbackQuery
	BotLang       string
	UserID        int
	UserLang      string
	Command       string
	Params        Parameters
	Err           error
}

type Parameters struct {
	ReplyText string
	Level     string
	Partition string
	Link      *LinkInfo
}

type LinkInfo struct {
	Url      string
	FileID   string
	Duration int
}

func SendAdminListMenu(botLang string, callback *tgbotapi.CallbackQuery) {
	s := &Situation{
		CallbackQuery: callback,
		BotLang:       botLang,
		UserID:        callback.From.ID,
	}

	if strings.Contains(callback.Data, "/") {
		adminSettingLevel(s)
		return
	}

	lang := assets.AdminLang(callback.From.ID)
	text := assets.AdminText(lang, "admin_list_text")

	markUp := msgs2.NewIlMarkUp(
		msgs2.NewIlRow(msgs2.NewIlAdminButton("add_admin_button", "admin/admin_setting/admin_list/add_admin_msg")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("delete_admin_button", "admin/admin_setting/admin_list/delete_admin")),
		msgs2.NewIlRow(msgs2.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	sendMsgAdnAnswerCallback(s, &markUp, text)
}

func sendMsgAdnAnswerCallback(s *Situation, markUp *tgbotapi.InlineKeyboardMarkup, text string) {
	if db.RdbGetAdminMsgID(s.BotLang, s.UserID) != 0 {
		msgs2.NewEditMarkUpMessage(s.BotLang, s.UserID, db.RdbGetAdminMsgID(s.BotLang, s.UserID), markUp, text)
		return
	}
	msgID := msgs2.NewIDParseMarkUpMessage(s.BotLang, int64(s.UserID), markUp, text)
	db.RdbSetAdminMsgID(s.BotLang, s.UserID, msgID)

	if s.CallbackQuery != nil {
		if s.CallbackQuery.ID != "" {
			msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
		}
	}
}

func adminSettingLevel(s *Situation) {
	fmt.Println(s.BotLang, s.CallbackQuery.Data)
}
