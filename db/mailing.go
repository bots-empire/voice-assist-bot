package db

import (
	"fmt"

	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

const (
	getLangIDQuery = "SELECT id, lang FROM users ORDER BY id LIMIT ? OFFSET ?;"
)

var (
	message           = make(map[string]tgbotapi.MessageConfig, 10)
	usersPerIteration = 100
)

func StartMailing(botLang string, initiator *model.User) {
	fillMessageMap()

	var (
		sendToUsers  int
		blockedUsers int
	)

	msgs.SendNotificationToDeveloper(
		fmt.Sprintf("%s // mailing started", botLang),
	)

	for offset := 0; ; offset += usersPerIteration {
		countSend, errCount := mailToUserWithPagination(botLang, offset)
		if countSend == -1 {
			sendRespMsgToMailingInitiator(botLang, initiator, "failing_mailing_text", sendToUsers)
			break
		}

		if countSend == 0 && errCount == 0 {
			break
		}

		sendToUsers += countSend
		blockedUsers += errCount
	}

	msgs.SendNotificationToDeveloper(
		fmt.Sprintf("%s // send to %d users mail", botLang, sendToUsers),
	)

	sendRespMsgToMailingInitiator(botLang, initiator, "complete_mailing_text", sendToUsers)

	assets.AdminSettings.BlockedUsers[botLang] = blockedUsers
	assets.SaveAdminSettings()
}

func sendRespMsgToMailingInitiator(botLang string, user *model.User, key string, countOfSends int) {
	lang := assets.AdminLang(user.ID)
	text := fmt.Sprintf(assets.AdminText(lang, key), countOfSends)

	_ = msgs.NewParseMessage(botLang, user.ID, text)
}

func mailToUserWithPagination(botLang string, offset int) (int, int) {
	users, err := getUsersWithPagination(botLang, offset)
	if err != nil {
		msgs.SendNotificationToDeveloper(errors.Wrap(err, "get users with pagination").Error())
		return -1, 0
	}

	totalCount := len(users)
	if totalCount == 0 {
		return 0, 0
	}

	responseChan := make(chan bool)
	var sendToUsers int

	for _, user := range users {
		go sendMailToUser(botLang, user, responseChan)
	}

	for countOfResp := 0; countOfResp < len(users); countOfResp++ {
		select {
		case resp := <-responseChan:
			if resp {
				sendToUsers++
			}
		}
	}

	return sendToUsers, totalCount - sendToUsers
}

func getUsersWithPagination(botLang string, offset int) ([]*model.User, error) {
	rows, err := model.GetDB(botLang).Query(getLangIDQuery, usersPerIteration, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed execute query")
	}

	var users []*model.User

	for rows.Next() {
		var (
			id   int64
			lang string
		)

		if err := rows.Scan(&id, &lang); err != nil {
			return nil, errors.Wrap(err, "failed scan row")
		}

		if containsInAdmin(id) {
			continue
		}

		users = append(users, &model.User{
			ID:       id,
			Language: lang,
		})
	}

	return users, nil
}

func sendMailToUser(botLang string, user *model.User, respChan chan<- bool) {
	msg := message[user.Language]
	msg.ChatID = user.ID

	_, err := model.GetGlobalBot(botLang).Bot.Send(msg)
	respChan <- err == nil
}

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
