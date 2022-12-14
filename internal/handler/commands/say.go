package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strings"
)

func (h BotCommandHelp) Say(request BotCommand) (response HelpResponse) {
	response.Help = "Give Bender a line of text to say in a channel. " +
		"Usage: '!say in {channel} {text}' or '!say {text}' for same channel."

	response.Description = "Cause the bot to say something in a channel. Usage: !say {text}"

	return response
}

func (bc BotCommand) Say(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Message:        fmt.Sprintf(`/echo "%s" 1`, event.body),
		Type:           "command",
		UserId:         u.Id,
	}

	event.ResponseChannel <- response

	return nil
}
