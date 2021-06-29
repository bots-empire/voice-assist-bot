package admin

import (
	"database/sql"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/msgs"
	"log"
)

func countUsers(botLang string) int {
	dataBase := assets.GetDB(botLang)
	rows, err := dataBase.Query("SELECT COUNT(*) FROM users;")
	if err != nil {
		text := "Fatal Err with DB - count_users.14 //" + err.Error()
		msgs.NewParseMessage("it", 1418862576, text)
		panic(err.Error())
	}

	return readRows(rows)
}

func readRows(rows *sql.Rows) int {
	defer rows.Close()

	var count int

	for rows.Next() {
		if err := rows.Scan(&count); err != nil {
			panic("Failed to scan row: " + err.Error())
		}
	}

	return count
}

func countAllUsers() int {
	var sum int
	for _, handler := range assets.Bots {
		rows, err := handler.DataBase.Query("SELECT COUNT(*) FROM users;")
		if err != nil {
			log.Println(err.Error())
		}
		sum += readRows(rows)
	}
	return sum
}

func countBlockedUsers(botLang string) int {
	//var count int
	//for _, value := range assets.AdminSettings.BlockedUsers {
	//	count += value
	//}
	//return count
	return assets.AdminSettings.BlockedUsers[botLang]
}

func countSubscribers(botLang string) int {
	rows, err := assets.Bots[botLang].DataBase.Query("SELECT COUNT(DISTINCT id) FROM subs;")
	if err != nil {
		log.Println(err.Error())
	}

	return readRows(rows)
}
