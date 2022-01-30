package db

import (
	"database/sql"
	"fmt"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	getLangIDQuery = "SELECT id, lang FROM users;"
)

var (
	message = make(map[string]tgbotapi.MessageConfig, 10)
)

func StartMailing(botLang string) {
	dataBase := model.Bots[botLang].DataBase
	rows, err := dataBase.Query(getLangIDQuery)
	if err != nil {
		panic(err.Error())
	}

	MailToUser(botLang, rows)
}

func MailToUser(botLang string, rows *sql.Rows) {
	defer rows.Close()
	fillMessageMap()

	var (
		sendToUsers  int
		blockedUsers int
	)

	for rows.Next() {
		var (
			id   int64
			lang string
		)

		if err := rows.Scan(&id, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		if containsInAdmin(id) {
			continue
		}

		msg := message[lang]
		msg.ChatID = id

		if !msgs.SendMessageToChat(botLang, msg) {
			blockedUsers += 1
			continue
		}

		sendToUsers++
	}

	_ = msgs.SendMsgToUser("it", tgbotapi.NewMessage(1418862576,
		fmt.Sprintf("%s // send to %d users mail", botLang, sendToUsers)),
	)

	assets.AdminSettings.BlockedUsers[botLang] = blockedUsers
	assets.SaveAdminSettings()
}

//func MailToUser(botLang string, rows *sql.Rows) {
//	fillMessageMap()
//
//	var users []*model.User
//
//	for rows.Next() {
//		var (
//			id   int64
//			lang string
//		)
//
//		if err := rows.Scan(&id, &lang); err != nil {
//			panic("Failed to scan row: " + err.Error())
//		}
//
//		if containsInAdmin(id) {
//			continue
//		}
//
//		users = append(users, &model.User{
//			ID:       id,
//			Language: lang,
//		})
//	}
//	rows.Close()
//
//	var blockedUsers int
//	mu := &sync.Mutex{}
//
//	for _, user := range users {
//		msg := message[user.Language]
//		msg.ChatID = user.ID
//
//		go func(config tgbotapi.MessageConfig) {
//			if !msgs.SendMessageToChat(botLang, config) {
//				mu.Lock()
//				blockedUsers += 1
//				mu.Unlock()
//			}
//		}(msg)
//	}
//
//	assets.AdminSettings.BlockedUsers[botLang] = blockedUsers
//	assets.SaveAdminSettings()
//}

func containsInAdmin(userID int64) bool {
	for key := range assets.AdminSettings.AdminID {
		if key == userID {
			return true
		}
	}
	return false
}

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
