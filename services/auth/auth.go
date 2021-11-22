package auth

import (
	"database/sql"
	"fmt"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"strconv"
	"strings"
)

const (
	typeFriend = "friend"
	//typeGroup  = "group"

	newUserQuery             = "INSERT INTO users VALUES(?, ?, 0, 0, FALSE, ?, ?, ?, FALSE);"
	UpdateAfterReferralQuery = "UPDATE users SET balance = ?, referral_%s = ? WHERE id = ?;"
	getUsersUserQuery        = "SELECT * FROM users WHERE id = ?;"
)

type ParentOfRef struct {
	ID        int
	TypeOfRef string
}

func CheckingTheUser(botLang string, message *tgbotapi.Message) (*model.User, error) {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query(getUsersUserQuery, message.From.ID)
	if err != nil {
		return nil, errors.Wrap(err, "get user")
	}

	users, err := ReadUsers(rows)
	if err != nil {
		return nil, errors.Wrap(err, "read user")
	}

	switch len(users) {
	case 0:
		user := createSimpleUser(botLang, message)
		if user.Language != "not_defined" {
			referralID := pullReferralID(message)
			if err := addNewUser(message, user, botLang, referralID); err != nil {
				return nil, errors.Wrap(err, "add new user")
			}
		} else {
			return user, model.ErrNotSelectedLanguage
		}
		return user, nil
	case 1:
		return users[0], nil
	default:
		return nil, model.ErrFoundTwoUsers
	}
}

func addNewUser(message *tgbotapi.Message, u *model.User, botLang string, referralID int64) error {
	fmt.Println(u.Language)
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query("INSERT INTO users VALUES(?, 0, 0, 0, 0, 0, FALSE, ?);", u.ID, u.Language)
	if err != nil {
		text := "Fatal Err with DB - auth.70 //" + err.Error()
		//msgs.NewParseMessage("it", 1418862576, text)
		log.Println(text)
		return errors.Wrap(err, "query failed")
	}
	rows.Close()

	if referralID == u.ID || referralID == 0 {
		return nil
	}

	baseUser, err := GetUser(botLang, referralID)
	if err != nil {
		errors.Wrap(err, "get user")
	}
	baseUser.Balance += assets.AdminSettings.Parameters[botLang].ReferralAmount
	rows, err = dataBase.Query("UPDATE users SET balance = ?, referral_count = ? WHERE id = ?;",
		baseUser.Balance, baseUser.ReferralCount+1, baseUser.ID)
	if err != nil {
		text := "Fatal Err with DB - auth.85 //" + err.Error()
		msgs.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}
	rows.Close()

	return nil
}

func pullReferralID(message *tgbotapi.Message) int64 {
	str := strings.Split(message.Text, " ")
	if len(str) < 2 {
		return 0
	}

	id, err := strconv.Atoi(str[1])
	if err != nil {
		log.Println(err)
		return 0
	}

	if id > 0 {
		return int64(id)
	}
	return 0
}

func createSimpleUser(botLang string, message *tgbotapi.Message) *model.User {
	return &model.User{
		ID:       message.From.ID,
		Language: botLang,
	}
}

func GetUser(botLang string, id int64) (*model.User, error) {
	dataBase := model.GetDB(botLang)
	rows, err := dataBase.Query(getUsersUserQuery, id)
	if err != nil {
		return nil, err
	}

	users, err := ReadUsers(rows)
	if err != nil || len(users) == 0 {
		return nil, model.ErrUserNotFound
	}
	return users[0], nil
}

func ReadUsers(rows *sql.Rows) ([]*model.User, error) {
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var (
			id                                                int64
			balance, completed, completedToday, referralCount int
			lastVoice                                         int64
			takeBonus                                         bool
			lang                                              string
		)

		if err := rows.Scan(&id, &balance, &completed, &completedToday, &lastVoice, &referralCount, &takeBonus, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, &model.User{
			ID:             id,
			Balance:        balance,
			Completed:      completed,
			CompletedToday: completedToday,
			LastVoice:      lastVoice,
			ReferralCount:  referralCount,
			TakeBonus:      takeBonus,
			Language:       lang,
		})
	}

	return users, nil
}
