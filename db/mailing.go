package db

import (
	"database/sql"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var (
	message = make(map[string]tgbotapi.MessageConfig, 5)
)

func StartMailing(botLang string) {
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("SELECT id, lang FROM users;")
	if err != nil {
		//text := "Fatal Err with DB - mailing.18 //" + err.Error()
		//msgs.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}

	MailToUsers(botLang, rows)
}

func MailToUsers(botLang string, rows *sql.Rows) {
	defer rows.Close()
	fillMessageMap()

	//blockedUsers := copyBlockedMap()
	//clearSelectedLang(blockedUsers)

	for rows.Next() {
		var (
			id   int
			lang string
		)

		if err := rows.Scan(&id, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		msg := message[lang]
		msg.ChatID = int64(id)

		if containsInAdmin(id) {
			continue
		}

		if !msgs.SendMessageToChat(botLang, msg) {
			assets.AdminSettings.BlockedUsers[botLang] += 1
		}
	}

	//assets.AdminSettings.BlockedUsers = blockedUsers
	assets.SaveAdminSettings()
}

func copyBlockedMap() map[string]int {
	blockedUsers := make(map[string]int, 5)
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			blockedUsers[lang] = 0
		}
	}
	return blockedUsers
}

func clearSelectedLang(blockedUsers map[string]int) {
	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			blockedUsers[lang] = 0
		}
	}
}

func containsInAdmin(userID int) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

//func createAStringOfLang() string {
//	var str string
//
//	for _, lang := range assets.AvailableLang {
//		if assets.AdminSettings.LangSelectedMap[lang] {
//			str += " lang = '" + lang + "' OR"
//		}
//	}
//	return strings.TrimRight(str, " OR")
//}

func fillMessageMap() {
	for _, lang := range assets.AvailableLang {
		text := assets.AdminSettings.AdvertisingText[lang]

		markUp := msgs.NewIlMarkUp(
			msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingChan[lang].Url)),
		).Build(lang)

		message[lang] = tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ReplyMarkup: markUp,
			},
			Text: text,
		}
	}
}
