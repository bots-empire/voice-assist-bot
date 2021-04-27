package regexp

import "regexp"

var (
	InvitationLink = regexp.MustCompile("^https://t.me/joinchat/[A-z0-9-]{1,30}$")
)
