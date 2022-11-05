package commands

import (
	"fmt"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Message(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "command",
	}

	channelObj, _ := event.mm.GetChannel(event.mm.DebuggingChannel.Name)
	response.ReplyChannelId = channelObj.Id

	response.Message = fmt.Sprintf(`/msg %s %s`, u.Name, strings.TrimLeft(event.body, u.Name))

	event.ResponseChannel <- response
	return nil
}
