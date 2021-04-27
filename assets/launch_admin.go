package assets

import (
	"encoding/json"
	"fmt"
	"os"
)

type Admin struct {
	AdminID             map[int]*AdminUser
	BonusAmount         int
	MinWithdrawalAmount int
	VoiceAmount         int
	MaxOfVoicePerDay    int
	ReferralAmount      int
	AdvertisingURL      string
	BlockedUsers        map[string]int
	LangSelectedMap     map[string]bool
	AdvertisingText     map[string]string
}

type AdminUser struct {
	Language  string
	FirstName string
}

var AdminSettings *Admin

func UploadAdminSettings() {
	var settings *Admin
	data, err := os.ReadFile("assets/admin.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		fmt.Println(err)
	}

	AdminSettings = settings
}

func SaveAdminSettings() {
	data, err := json.MarshalIndent(AdminSettings, "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile("assets/admin.json", data, 0600); err != nil {
		panic(err)
	}
}
