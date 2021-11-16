package assets

import (
	"encoding/json"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/Stepan1328/voice-assist-bot/model"
	"math/rand"
	"os"
	"strings"
)

const (
	commandsPath             = "assets/commands"
	beginningOfAdminLangPath = "assets/admin/"
	beginningOfUserLangPath  = "assets/language/"
	beginningOfSiriLangPath  = "assets/siri/"
)

var (
	AvailableAdminLang = []string{"en", "ru"}
	AvailableLang      = []string{"it", "pt", "es", "mx", "ch"}

	Commands     = make(map[string]string)
	Language     = make([]map[string]string, len(AvailableLang))
	AdminLibrary = make([]map[string]string, 2)
	Task         = make([][]string, len(AvailableLang))
)

func ParseLangMap() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile(beginningOfUserLangPath + lang + jsonFormatName)
		_ = json.Unmarshal(bytes, &Language[i])
	}
}

func LangText(lang, key string) string {
	index := findLangIndex(lang)
	return Language[index][key]
}

func ParseCommandsList() {
	bytes, _ := os.ReadFile(commandsPath + jsonFormatName)
	_ = json.Unmarshal(bytes, &Commands)
}

func ParseSiriTasks() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile(beginningOfSiriLangPath + lang + jsonFormatName)
		_ = json.Unmarshal(bytes, &Task[i])
	}
}

func GetCommandFromText(message *tgbotapi.Message, userLang string, userID int64) (string, error) {
	searchText := getSearchText(message)
	for key, text := range Language[findLangIndex(userLang)] {
		if text == searchText {
			return Commands[key], nil
		}
	}

	if command := searchInCommandAdmins(userID, searchText); command != "" {
		return command, nil
	}

	command := Commands[searchText]
	if command != "" {
		return command, nil
	}

	return "", model.ErrCommandNotConverted
}

func getSearchText(message *tgbotapi.Message) string {
	if message.Command() != "" {
		return strings.Split(message.Text, " ")[0]
	}
	return message.Text
}

func searchInCommandAdmins(userID int64, searchText string) string {
	lang := getAdminLang(userID)
	for key, text := range AdminLibrary[findAdminLangIndex(lang)] {
		if text == searchText {
			return Commands[key]
		}
	}
	return ""
}

func getAdminLang(userID int64) string {
	for key := range AdminSettings.AdminID {
		if key == userID {
			return AdminSettings.AdminID[key].Language
		}
	}
	return ""
}

func SiriText(lang string) string {
	index := findLangIndex(lang)
	num := rand.Intn(len(Task[index]))
	return Task[index][num]
}

func ParseAdminMap() {
	for i, lang := range AvailableAdminLang {
		bytes, _ := os.ReadFile(beginningOfAdminLangPath + lang + jsonFormatName)
		_ = json.Unmarshal(bytes, &AdminLibrary[i])
	}
}

func AdminText(lang, key string) string {
	index := findAdminLangIndex(lang)
	return AdminLibrary[index][key]
}

func findLangIndex(lang string) int {
	for i, elem := range AvailableLang {
		if elem == lang {
			return i
		}
	}
	return 0
}

func findAdminLangIndex(lang string) int {
	for i, elem := range AvailableAdminLang {
		if elem == lang {
			return i
		}
	}
	return 0
}

func AdminLang(userID int64) string {
	return AdminSettings.AdminID[userID].Language
}
