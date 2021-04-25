package db

import (
	"github.com/Stepan1328/voice-assist-bot/assets"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"strconv"
)

func RdbSetUser(ID int, level string) {
	userID := userIDToRdb(ID)
	_, err := Rdb.Set(userID, level, 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func userIDToRdb(userID int) string {
	return "user:" + strconv.Itoa(userID)
}

func GetLevel(id int) string {
	userID := userIDToRdb(id)
	have, err := Rdb.Exists(userID).Result()
	if err != nil {
		log.Println(err)
	}
	if have == 0 {
		return "empty"
	}

	value, err := Rdb.Get(userID).Result()
	if err != nil {
		log.Println(err)
	}
	return value
}

func RdbSetTemporary(userID, msgID int) {
	temporaryID := temporaryIDToRdb(userID)
	_, err := Rdb.Set(temporaryID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func temporaryIDToRdb(userID int) string {
	return "message:" + strconv.Itoa(userID)
}

func RdbGetTemporary(userID int) string {
	temporaryID := temporaryIDToRdb(userID)
	result, err := Rdb.Get(temporaryID).Result()
	if err != nil {
		log.Println(err)
	}
	return result
}

func RdbSetAdminMsgID(userID, msgID int) {
	adminMsgID := adminMsgIDToRdb(userID)
	_, err := Rdb.Set(adminMsgID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func adminMsgIDToRdb(userID int) string {
	return "admin_msg_id:" + strconv.Itoa(userID)
}

func RdbGetAdminMsgID(userID int) int {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}
	msgID, _ := strconv.Atoi(result)
	return msgID
}

func DeleteOldAdminMsg(userID int) {
	adminMsgID := adminMsgIDToRdb(userID)
	result, err := Rdb.Get(adminMsgID).Result()
	if err != nil {
		log.Println(err)
	}

	if oldMsgID, _ := strconv.Atoi(result); oldMsgID != 0 {
		msg := tgbotapi.NewDeleteMessage(int64(userID), oldMsgID)
		if _, err = assets.Bot.Send(msg); err != nil && err.Error() != "message to delete not found" {
			log.Println(err)
		}
		RdbSetAdminMsgID(userID, 0)
	}
}
