package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Apb(event BotCommand) (response Response, err error) {
	response.Type = "command"
	u, err := users.HasUser(event.body, event.cache)
	if u {
		response.Message = fmt.Sprintf(`/me sends out the blood hounds to find %s`, event.body)
	} else {
		response.Type = "dm"
		response.Message = fmt.Sprintf(`Who's ` + event.body + `?`)
	}

	return response, err
}
