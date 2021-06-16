package admin

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/Stepan1328/voice-assist-bot/db"
	msgs2 "github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	AvailableSymbolInKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	AdminKeyLength       = 24
	LinkLifeTime         = 180
	GodUserID            = 138814168
)

var availableKeys = make(map[string]string)

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

func CheckNewAdmin(s Situation) {
	key := strings.Replace(s.Command, "/start new_admin_", "", 1)
	if availableKeys[key] != "" {
		assets.AdminSettings.AdminID[s.UserID] = &assets.AdminUser{
			Language:  "ru",
			FirstName: s.Message.From.FirstName,
		}
		if s.UserID == GodUserID {
			assets.AdminSettings.AdminID[s.UserID].SpecialPossibility = true
		}
		assets.SaveAdminSettings()

		text := assets.AdminText(s.UserLang, "welcome_to_admin")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		availableKeys[key] = ""
		return
	}

	text := assets.LangText(s.UserLang, "invalid_link_err")
	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
}

func adminSettingLevel(s *Situation) {
	s.CallbackQuery.Data = strings.Replace(s.CallbackQuery.Data, "admin_list/", "", 1)

	switch s.CallbackQuery.Data {
	case "add_admin_msg":
		NewAdminToListMsg(s)
	case "delete_admin":
		DeleteAdminCommand(s)
	}
}

func NewAdminToListMsg(s *Situation) {
	lang := assets.AdminLang(s.UserID)

	link := createNewAdminLink(s.BotLang)
	text := adminFormatText(lang, "new_admin_key_text", link, LinkLifeTime)

	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	SendAdminListMenu(s.BotLang, s.CallbackQuery)
	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

func createNewAdminLink(botLang string) string {
	key := generateKey()
	availableKeys[key] = key
	go deleteKey(key)
	return cfg.GetBotConfig(botLang).Link + "?start=new_admin_" + key
}

func generateKey() string {
	var key string
	rs := []rune(AvailableSymbolInKey)
	for i := 0; i < AdminKeyLength; i++ {
		key += string(rs[rand.Intn(len(AvailableSymbolInKey))])
	}
	return key
}

func deleteKey(key string) {
	time.Sleep(time.Second * LinkLifeTime)
	availableKeys[key] = ""
}

func DeleteAdminCommand(s *Situation) {
	if !adminHavePrivileges(s) {
		msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "admin_dont_have_permissions")
		return
	}

	lang := assets.AdminLang(s.UserID)
	db.RdbSetUser(s.BotLang, s.UserID, "admin/"+s.CallbackQuery.Data)

	msgs2.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	msgs2.NewParseMessage(s.BotLang, int64(s.UserID), createListOfAdminText(lang))
}

func adminHavePrivileges(s *Situation) bool {
	return assets.AdminSettings.AdminID[s.UserID].SpecialPossibility
}

func createListOfAdminText(lang string) string {
	var listOfAdmins string
	for i, admin := range assets.AdminSettings.AdminID {
		listOfAdmins += strconv.Itoa(i+1) + ") " + admin.FirstName + "\n"
		//listOfAdmins += "Language: " + admin.Language + "\nSpecial possibility: "
		//if admin.SpecialPossibility {
		//	listOfAdmins += "yes\n\n"
		//} else {
		//	listOfAdmins += "no\n\n"
		//}
	}

	return adminFormatText(lang, "delete_admin_body_text", listOfAdmins)
}

func RemoveAdminCommand(botLang string, message *tgbotapi.Message) {
	s := &Situation{
		Message: message,
		BotLang: botLang,
		UserID:  message.From.ID,
	}

	lang := assets.AdminLang(s.UserID)
	adminId, err := strconv.Atoi(s.Message.Text)
	if err != nil {
		text := assets.AdminText(lang, "incorrect_admin_id_text")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	if !checkAdminIDInTheList(adminId) {
		text := assets.AdminText(lang, "incorrect_admin_id_text")
		msgs2.NewParseMessage(s.BotLang, int64(s.UserID), text)
		return
	}

	delete(assets.AdminSettings.AdminID, adminId)
	assets.SaveAdminSettings()
	setAdminBackButton(s.BotLang, s.UserID, "admin_removed_status")
	db.DeleteOldAdminMsg(s.BotLang, s.UserID)

	text := assets.LangText(lang, "main_select_menu")
	sendMenu(s.BotLang, s.UserID, text)
}

func checkAdminIDInTheList(adminID int) bool {
	_, inMap := assets.AdminSettings.AdminID[adminID]
	return inMap
}
