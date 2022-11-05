package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Apb(event BotCommand) error {
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
	}
	response.Type = "command"
	_, ok, err := users.GetUser(event.body, event.cache)
	if ok {
		response.Message = fmt.Sprintf(`/me sends out the blood hounds to find %s`, event.body)
	} else {
		response.Type = "dm"
		response.Message = fmt.Sprintf(`Who's ` + event.body + `?`)
	}

	event.ResponseChannel <- response

	return err
}
