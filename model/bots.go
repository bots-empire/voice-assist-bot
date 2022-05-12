package model

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Stepan1328/voice-assist-bot/cfg"

	_ "github.com/go-sql-driver/mysql"
)

const (
	tokensPath       = "./cfg/tokens.json"
	dbDriver         = "mysql"
	redisDefaultAddr = "127.0.0.1:6379"
)

var Bots = make(map[string]*GlobalBot)

type GlobalBot struct {
	BotLang string

	Bot      *tgbotapi.BotAPI
	Chanel   tgbotapi.UpdatesChannel
	Rdb      *redis.Client
	DataBase *sql.DB

	MessageHandler  GlobalHandlers
	CallbackHandler GlobalHandlers

	AdminMessageHandler  GlobalHandlers
	AdminCallBackHandler GlobalHandlers

	Commands     map[string]string
	Language     map[string]map[string]string
	AdminLibrary map[string]map[string]string
	Tasks        map[string][]string

	BotToken      string   `json:"bot_token"`
	BotLink       string   `json:"bot_link"`
	LanguageInBot []string `json:"language_in_bot"`
	AssistName    string   `json:"assist_name"`

	MaintenanceMode bool
}

type GlobalHandlers interface {
	GetHandler(command string) Handler
}

type Handler func(situation *Situation) error

func UploadDataBase(dbLang string) *sql.DB {
	dataBase, err := sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/") //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	dataBase.Exec("CREATE DATABASE IF NOT EXISTS " + cfg.DBCfg.Names[dbLang] + ";")
	dataBase.Exec("USE " + cfg.DBCfg.Names[dbLang] + ";")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS users (" + cfg.UserTable + ");")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS links (" + cfg.Links + ");")
	dataBase.Exec("CREATE TABLE IF NOT EXISTS subs (" + cfg.Subs + ");")

	dataBase.Close()

	dataBase, err = sql.Open(dbDriver, cfg.DBCfg.User+cfg.DBCfg.Password+"@/"+cfg.DBCfg.Names[dbLang]) //TODO: refactor
	if err != nil {
		log.Fatalf("Failed open database: %s\n", err.Error())
	}

	dataBase.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS advert_channel int NOT NULL AFTER last_voice;")

	TakeAllUsers(dataBase)

	err = dataBase.Ping()
	if err != nil {
		log.Fatalf("Failed upload database: %s\n", err.Error())
	}

	return dataBase
}

func TakeAllUsers(dataBase *sql.DB) {
	rand.Seed(time.Now().Unix())
	rows, err := dataBase.Query(`SELECT * FROM users WHERE advert_channel = 0;`)
	if err != nil {
	}

	if rows == nil {
		return
	}

	users, err := ReadUser(rows)
	if err != nil {
	}

	for i := range users {
		dataBase.Exec(`UPDATE users SET advert_channel = ? WHERE id = ?;`, rand.Intn(3)+1, users[i].ID)
	}
}

func ReadUser(rows *sql.Rows) ([]*User, error) {
	defer rows.Close()

	var users []*User

	for rows.Next() {
		user := &User{}

		if err := rows.Scan(&user.ID,
			&user.Balance,
			&user.Completed,
			&user.CompletedToday,
			user.LastVoice,
			&user.AdvertChannel,
			&user.ReferralCount,
			&user.TakeBonus,
			&user.Language,
		); err != nil {
			//msgs.SendNotificationToDeveloper(errors.Wrap(err, "failed to scan row").Error())
		}

		users = append(users, user)
	}

	return users, nil
}

func StartRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisDefaultAddr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return rdb
}

func GetDB(botLang string) *sql.DB {
	return Bots[botLang].DataBase
}

func FillBotsConfig() {
	bytes, err := os.ReadFile(tokensPath)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(bytes, &Bots)
	if err != nil {
		panic(err)
	}

	for lang, bot := range Bots {
		bot.BotLang = lang
	}

	fmt.Println(Bots)
}

func GetGlobalBot(botLang string) *GlobalBot {
	return Bots[botLang]
}

func (b *GlobalBot) GetBot() *tgbotapi.BotAPI {
	return b.Bot
}

func (b *GlobalBot) GetDataBase() *sql.DB {
	return b.DataBase
}

func (b *GlobalBot) AvailableLang() []string {
	return b.LanguageInBot
}

func (b *GlobalBot) GetCurrency() string {
	return AdminSettings.GetCurrency(b.BotLang)
}

func (b *GlobalBot) LangText(lang, key string, values ...interface{}) string {
	formatText := b.Language[lang][key]
	return fmt.Sprintf(formatText, values...)
}

func (b *GlobalBot) GetTexts(lang string) map[string]string {
	return b.Language[lang]
}

func (b *GlobalBot) CheckAdmin(userID int64) bool {
	_, exist := AdminSettings.AdminID[userID]
	return exist
}

func (b *GlobalBot) AdminLang(userID int64) string {
	return AdminSettings.AdminID[userID].Language
}

func (b *GlobalBot) AdminText(adminLang, key string) string {
	return b.AdminLibrary[adminLang][key]
}

func (b *GlobalBot) UpdateBlockedUsers(channel int) {
}

func (b *GlobalBot) GetAdvertURL(userLang string, channel int) string {
	return AdminSettings.GetAdvertUrl(userLang, channel)
}

func (b *GlobalBot) GetAdvertText(userLang string, channel int) string {
	return AdminSettings.GetAdvertText(userLang, channel)
}

func (b *GlobalBot) GetAdvertisingPhoto(lang string, channel int) string {
	return AdminSettings.GlobalParameters[lang].AdvertisingPhoto[channel]
}

func (b *GlobalBot) GetAdvertisingVideo(lang string, channel int) string {
	return AdminSettings.GlobalParameters[lang].AdvertisingVideo[channel]
}

func (b *GlobalBot) ButtonUnderAdvert() bool {
	return AdminSettings.GlobalParameters[b.BotLang].Parameters.ButtonUnderAdvert
}

func (b *GlobalBot) AdvertisingChoice(channel int) string {
	return AdminSettings.GlobalParameters[b.BotLang].AdvertisingChoice[channel]
}
