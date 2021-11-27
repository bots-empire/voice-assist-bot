package administrator

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const (
	AvailableSymbolInKey = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	AdminKeyLength       = 24
	LinkLifeTime         = 180
	GodUserID            = 872383555
)

var availableKeys = make(map[string]string)

type AdminListCommand struct {
}

func NewAdminListCommand() *AdminListCommand {
	return &AdminListCommand{}
}

func (c *AdminListCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)
	text := assets.AdminText(lang, "admin_list_text")

	markUp := msgs.NewIlMarkUp(
		msgs.NewIlRow(msgs.NewIlAdminButton("add_admin_button", "admin/add_admin_msg")),
		msgs.NewIlRow(msgs.NewIlAdminButton("delete_admin_button", "admin/delete_admin")),
		msgs.NewIlRow(msgs.NewIlAdminButton("back_to_admin_settings", "admin/admin_setting")),
	).Build(lang)

	return sendMsgAdnAnswerCallback(s, &markUp, text)
}

func CheckNewAdmin(s model.Situation) error {
	key := strings.Replace(s.Command, "/start new_admin_", "", 1)
	if availableKeys[key] != "" {
		assets.AdminSettings.AdminID[s.User.ID] = &assets.AdminUser{
			Language:  "ru",
			FirstName: s.Message.From.FirstName,
		}
		if s.User.ID == GodUserID {
			assets.AdminSettings.AdminID[s.User.ID].SpecialPossibility = true
		}
		assets.SaveAdminSettings()

		text := assets.AdminText(s.User.Language, "welcome_to_admin")
		delete(availableKeys, key)
		return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	}

	text := assets.LangText(s.User.Language, "invalid_link_err")
	return msgs.NewParseMessage(s.BotLang, s.User.ID, text)
}

type NewAdminToListCommand struct {
}

func NewNewAdminToListCommand() *NewAdminToListCommand {
	return &NewAdminToListCommand{}
}

func (c *NewAdminToListCommand) Serve(s model.Situation) error {
	lang := assets.AdminLang(s.User.ID)

	link := createNewAdminLink(s.BotLang)
	text := adminFormatText(lang, "new_admin_key_text", link, LinkLifeTime)

	err := msgs.NewParseMessage(s.BotLang, s.User.ID, text)
	if err != nil {
		return err
	}
	db.DeleteOldAdminMsg(s.BotLang, s.User.ID)
	s.Command = "/send_admin_list"
	if err := NewAdminListCommand().Serve(s); err != nil {
		return err
	}

	return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "make_a_choice")
}

func createNewAdminLink(botLang string) string {
	key := generateKey()
	availableKeys[key] = key
	go deleteKey(key)
	return model.GetGlobalBot(botLang).BotLink + "?start=new_admin_" + key
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

type DeleteAdminCommand struct {
}

func NewDeleteAdminCommand() *DeleteAdminCommand {
	return &DeleteAdminCommand{}
}

func (c *DeleteAdminCommand) Serve(s model.Situation) error {
	if !adminHavePrivileges(s) {
		return msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "admin_dont_have_permissions")
	}

	lang := assets.AdminLang(s.User.ID)
	db.RdbSetUser(s.BotLang, s.User.ID, s.CallbackQuery.Data)

	_ = msgs.SendAdminAnswerCallback(s.BotLang, s.CallbackQuery, "type_the_text")
	return msgs.NewParseMessage(s.BotLang, s.User.ID, createListOfAdminText(lang))
}

func adminHavePrivileges(s model.Situation) bool {
	return assets.AdminSettings.AdminID[s.User.ID].SpecialPossibility
}

func createListOfAdminText(lang string) string {
	var listOfAdmins string
	for i, admin := range assets.AdminSettings.AdminID {
		listOfAdmins += strconv.Itoa(int(i+1)) + ") " + admin.FirstName + "\n"
		//listOfAdmins += "Language: " + admin.Language + "\nSpecial possibility: "
		//if admin.SpecialPossibility {
		//	listOfAdmins += "yes\n\n"
		//} else {
		//	listOfAdmins += "no\n\n"
		//}
	}

	return adminFormatText(lang, "delete_admin_body_text", listOfAdmins)
}
