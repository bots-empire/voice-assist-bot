package cfg

type DBConfig struct {
	User     string
	Password string
	Names    map[string]string
}

//// DBCfg Slava local config
//var DBCfg = DBConfig{
//	User:     "root",
//	Password: ":!BlackR1",
//	Names: map[string]string{
//		"it":    "italy",
//		"pt":    "portugaly",
//		"es":    "espany",
//		"ar":    "argentina",
//		"ch":    "chile",
//		"mx":    "espany_2",
//		"fr":    "france",
//		"fr_en": "fr_en",
//	},
//}

// DBCfg Stepan local config
var DBCfg = DBConfig{
	User:     "root",
	Password: ":root",
	Names: map[string]string{
		"it":    "italy",
		"pt":    "portugaly",
		"es":    "espany",
		"ar":    "argentina",
		"ch":    "chile",
		"mx":    "espany_2",
		"fr":    "france",
		"fr_en": "fr_en",
	},
}
