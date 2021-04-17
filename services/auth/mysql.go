package auth

import (
	"database/sql"
	"github.com/Stepan1328/voice-assist-bot/db"
	_ "github.com/go-sql-driver/mysql"
)

func (user *User) CheckingTheUser() {
	rows, err := db.DataBase.Query("SELECT * FROM users WHERE id = ?;", user.ID)
	if err != nil {
		panic(err.Error())
	}

	users := ReadUsers(rows)

	switch len(users) {
	case 0:
		user.AddNewUser()
		users = append(users, *user)
	case 1:
	default:
		panic("There were two identical users")
	}
	*user = users[0]
}

func (user *User) AddNewUser() {
	_, err := db.DataBase.Query("INSERT INTO users VALUES(?, 0, 0, 0, 0, 0, ?);", user.ID, user.Language)
	if err != nil {
		panic(err.Error())
	}
}

func ReadUsers(rows *sql.Rows) []User {
	defer rows.Close()

	var users []User

	for rows.Next() {
		var (
			id, balance, completed, completedToday, referralCount int
			lastVoice                                             int64
			lang                                                  string
		)

		if err := rows.Scan(&id, &balance, &completed, &completedToday, &lastVoice, &referralCount, &lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			ID:             id,
			Balance:        balance,
			Completed:      completed,
			CompletedToday: completedToday,
			LastVoice:      lastVoice,
			ReferralCount:  referralCount,
			Language:       lang,
		})
	}

	return users
}

func GetLang(id int) string {
	rows, err := db.DataBase.Query("SELECT lang FROM users WHERE id = ?;", id)
	if err != nil {
		panic(err.Error())
	}

	return GetLangFromRow(rows)
}

func GetLangFromRow(rows *sql.Rows) string {
	defer rows.Close()

	var users []User

	for rows.Next() {
		var (
			lang string
		)

		if err := rows.Scan(&lang); err != nil {
			panic("Failed to scan row: " + err.Error())
		}

		users = append(users, User{
			Language: lang,
		})
	}

	if len(users) != 1 {
		panic("The number if users fond is not equal to one")
	}
	return users[0].Language
}
