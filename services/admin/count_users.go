package admin

import (
	"database/sql"
	"github.com/Stepan1328/voice-assist-bot/assets"
	"github.com/Stepan1328/voice-assist-bot/db"
	"log"
)

func countUsers() int {
	rows, err := db.DataBase.Query("SELECT COUNT(*) FROM users;")
	if err != nil {
		log.Println(err.Error())
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

func countBlockedUsers() int {
	var count int
	for _, value := range assets.AdminSettings.BlockedUsers {
		count += value
	}
	return count
}
