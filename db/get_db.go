package db

import (
	"database/sql"
	"github.com/go-redis/redis"
)

var DataBase *sql.DB

func UploadDataBase() {
	var err error
	DataBase, err = sql.Open("mysql", "root@/test")
	if err != nil {
		panic(err.Error())
	}
}

func StartRedis(k int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       k,  // use default DB
	})
	return rdb
}
