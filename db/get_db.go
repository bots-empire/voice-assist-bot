package db

import (
	"database/sql"
	"github.com/go-redis/redis"
)

var DataBase *sql.DB
var Rdb *redis.Client

func UploadDataBase() {
	var err error
	DataBase, err = sql.Open("mysql", "root@/test")
	if err != nil {
		panic(err.Error())
	}
}

func StartRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	Rdb = rdb
}
