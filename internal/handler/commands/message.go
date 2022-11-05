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

	target := strings.TrimLeft(event.target, "@")
	response.Message = fmt.Sprintf(`/msg %s %s`, target, event.body)

	event.ResponseChannel <- response
	return nil
}
