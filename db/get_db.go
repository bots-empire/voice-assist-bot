package db

import (
	"database/sql"
	"github.com/Stepan1328/voice-assist-bot/cfg"
	"github.com/go-redis/redis"
)

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open("mysql",
		cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang])
	if err != nil {
		panic(err.Error())
	}
	return dataBase
}

func StartRedis(k int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "", // no password set
		DB:       k,  // use default DB
	})
	return rdb
}
