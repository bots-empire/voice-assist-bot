package assets

import (
	"encoding/json"
	"fmt"
	"os"
)

type Admin struct {
	AdminID             []int
	BonusAmount         int
	MinWithdrawalAmount int
	VoiceAmount         int
	ReferralAmount      int
	AdvertisingURL      string
	TotalUsers          int
	ActiveUsers         int
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

	if err := os.WriteFile("assets/admin.json", data, 0600); err != nil {
		panic(err)
	}
}
