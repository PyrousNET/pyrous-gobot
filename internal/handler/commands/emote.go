package commands

import (
	"fmt"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Emote(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "command",
	}

	response.Message = fmt.Sprintf(`/me %s`, event.body)

	event.ResponseChannel <- response
	return nil
}
