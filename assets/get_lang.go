package assets

import (
	"encoding/json"
	"math/rand"
	"os"
)

var (
	AvailableLang      = []string{"en", "de", "it", "pt", "es"}
	AvailableLangIndex = map[string]int{"en": 0, "de": 1, "it": 2, "pt": 3, "es": 4}
	Language           = make([]map[string]string, 5)

	Task = make([][]string, 5)
)

func ParseLangMap() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile("./assets/language/" + lang + ".json")
		_ = json.Unmarshal(bytes, &Language[i])
	}
}

func LangText(lang, key string) string {
	index, _ := AvailableLangIndex[lang]
	return Language[index][key]
}

func ParseSiriTasks() {
	for i, lang := range AvailableLang {
		bytes, _ := os.ReadFile("./assets/siri/" + lang + ".json")
		_ = json.Unmarshal(bytes, &Task[i])
	}
}

func SiriText(lang string) string {
	index, _ := AvailableLangIndex[lang]
	num := rand.Intn(len(Task[index]))
	return Task[index][num]
}
