package assets

import (
	"encoding/json"
	"math/rand"
	"os"
)

var (
	AvailableAdminLang = []string{"en", "ru"}
	AvailableLang      = []string{"en", "de", "it", "pt", "es"}

	Language     = make([]map[string]string, 5)
	AdminLibrary = make([]map[string]string, 2)
	Task         = make([][]string, 5)
)

func ParseLangMap() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile("./assets/language/" + lang + ".json")
		_ = json.Unmarshal(bytes, &Language[i])
	}
}

func LangText(lang, key string) string {
	index := findLangIndex(lang)
	return Language[index][key]
}

func ParseSiriTasks() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile("./assets/siri/" + lang + ".json")
		_ = json.Unmarshal(bytes, &Task[i])
	}
}

func SiriText(lang string) string {
	index := findLangIndex(lang)
	num := rand.Intn(len(Task[index]))
	return Task[index][num]
}

func ParseAdminMap() {
	for i, lang := range AvailableAdminLang {
		bytes, _ := os.ReadFile("./assets/admin/" + lang + ".json")
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

func AdminLang(userID int) string {
	return AdminSettings.AdminID[userID].Language
}
