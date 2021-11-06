package cfg

import (
	"encoding/json"
	"os"
)

type BotConfig struct {
	Token         string
	Link          string
	LanguageInBot string

	StartLanguages []string
}

var Tokens = make(map[string]BotConfig)

func FillBotsConfig() {
	bytes, _ := os.ReadFile("./cfg/tokens.json")
	err := json.Unmarshal(bytes, &Tokens)
	if err != nil {
		panic(err)
	}
}

func GetBotConfig(botLang string) BotConfig {
	return Tokens[botLang]
}
