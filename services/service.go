package services

import (
	"github.com/Stepan1328/voice-assist-bot/model"
	"github.com/Stepan1328/voice-assist-bot/services/administrator"
	"github.com/Stepan1328/voice-assist-bot/services/auth"
	"github.com/bots-empire/base-bot/msgs"
)

type Users struct {
	bot *model.GlobalBot

	auth  *auth.Auth
	admin *administrator.Admin
	Msgs  *msgs.Service
}

func NewUsersService(bot *model.GlobalBot, auth *auth.Auth, admin *administrator.Admin, msgs *msgs.Service) *Users {
	return &Users{
		bot:   bot,
		auth:  auth,
		admin: admin,
		Msgs:  msgs,
	}
}
