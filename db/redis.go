package db

import (
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

func userIDToRdb(id int) string {
	userID := "user:" + strconv.Itoa(id)
	return userID
}

func RdbSetTemporary(ID, msgID int) {
	temporaryID := temporaryIDToRdb(ID)
	_, err := Rdb.Set(temporaryID, strconv.Itoa(msgID), 0).Result()
	if err != nil {
		log.Println(err)
	}
}

func temporaryIDToRdb(id int) string {
	temporaryID := "message:" + strconv.Itoa(id)
	return temporaryID
}

func RdbGetTemporary(userID int) string {
	temporaryID := temporaryIDToRdb(userID)
	result, err := Rdb.Get(temporaryID).Result()
	if err != nil {
		log.Println(err)
	}
	return result
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
