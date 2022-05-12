package model

import (
	"encoding/json"
	"math/rand"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	commandsPath             = "assets/commands"
	beginningOfAdminLangPath = "assets/admin/"
	beginningOfUserLangPath  = "assets/language/"
	beginningOfSiriLangPath  = "assets/siri/"
)

var (
	AvailableAdminLang = []string{"en", "ru"}
)

func (b *GlobalBot) ParseLangMap() {
	b.Language = make(map[string]map[string]string)

	for _, lang := range b.LanguageInBot {
		bytes, _ := os.ReadFile(beginningOfUserLangPath + lang + jsonFormatName)
		dictionary := make(map[string]string)

		_ = json.Unmarshal(bytes, &dictionary)
		b.Language[lang] = dictionary
	}
}

func (b *GlobalBot) ParseCommandsList() {
	bytes, _ := os.ReadFile(commandsPath + jsonFormatName)
	_ = json.Unmarshal(bytes, &b.Commands)
}

func (b *GlobalBot) ParseSiriTasks() {
	b.Tasks = make(map[string][]string)

	for _, lang := range b.LanguageInBot {
		bytes, _ := os.ReadFile(beginningOfSiriLangPath + lang + jsonFormatName)
		dictionary := make([]string, 0)

		_ = json.Unmarshal(bytes, &dictionary)
		b.Tasks[lang] = dictionary
	}
}

func (b *GlobalBot) GetCommandFromText(message *tgbotapi.Message, userLang string, userID int64) (string, error) {
	searchText := getSearchText(message)
	for key, text := range b.Language[userLang] {
		if text == searchText {
			return b.Commands[key], nil
		}
	}

	if command := b.searchInAdminCommands(userID, searchText); command != "" {
		return command, nil
	}

	command := b.Commands[searchText]
	if command != "" {
		return command, nil
	}

	return "", ErrCommandNotConverted
}

func getSearchText(message *tgbotapi.Message) string {
	if message.Command() != "" {
		return strings.Split(message.Text, " ")[0]
	}
	return message.Text
}

func (b *GlobalBot) searchInAdminCommands(userID int64, searchText string) string {
	lang := getAdminLang(userID)
	for key, text := range b.AdminLibrary[lang] {
		if text == searchText {
			return b.Commands[key]
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

func (b *GlobalBot) SiriText(lang string) string {
	num := rand.Intn(len(b.Tasks[lang]))
	return b.Tasks[lang][num]
}

func (b *GlobalBot) ParseAdminMap() {
	for _, lang := range AvailableAdminLang {
		bytes, _ := os.ReadFile(beginningOfAdminLangPath + lang + jsonFormatName)
		dictionary := make(map[string]string)

		_ = json.Unmarshal(bytes, &dictionary)
		b.AdminLibrary = make(map[string]map[string]string)
		b.AdminLibrary[lang] = dictionary
	}
}

func (b *GlobalBot) findLangIndex(lang string) int {
	for i, elem := range b.LanguageInBot {
		if elem == lang {
			return i
		}
	}
	return 0
}

func AdminLang(userID int64) string {
	return AdminSettings.AdminID[userID].Language
}
