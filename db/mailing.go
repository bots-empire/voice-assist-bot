package db

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"strings"
)

var (
	message = make(map[string]tgbotapi.MessageConfig, 5)
)

func StartMailing() {
	rows, err := DataBase.Query("SELECT id, lang FROM users WHERE" + createAStringOfLang() + ";")
	if err != nil {
		panic(err.Error())
	}

	MailToUser(rows)
}

func MailToUser(rows *sql.Rows) {
	defer rows.Close()
	fillMessageMap()

	blockedUsers := copyBlockedMap()
	clearSelectedLang(blockedUsers)

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
		fmt.Println(id, lang)

		if !msgs.SendMessageToChat(msg) {
			blockedUsers[lang] += 1
		}
	}

	assets.AdminSettings.BlockedUsers = blockedUsers
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

func createAStringOfLang() string {
	var str string

	for _, lang := range assets.AvailableLang {
		if assets.AdminSettings.LangSelectedMap[lang] {
			str += " lang = '" + lang + "' OR"
		}
	}
	return strings.TrimRight(str, " OR")
}

func fillMessageMap() {
	for _, lang := range assets.AvailableLang {
		text := assets.AdminSettings.AdvertisingText[lang]

		markUp := msgs.NewIlMarkUp(
			msgs.NewIlRow(msgs.NewIlURLButton("advertisement_button_text", assets.AdminSettings.AdvertisingURL)),
		).Build(lang)

		message[lang] = tgbotapi.MessageConfig{
			BaseChat: tgbotapi.BaseChat{
				ReplyMarkup: markUp,
			},
			Text: text,
		}
	}
}
