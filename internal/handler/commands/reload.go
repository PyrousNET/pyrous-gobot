package commands

import (
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strings"
)

func (h BotCommandHelp) Reload(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Causes the bot to shutdown, pull any changes from git, and restart",
		Description: "Reloads the bot",
	}
}

func (bc BotCommand) Reload(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "shutdown",
		Quit:           event.Quit,
	}

	response.Message = "So, if my body gets killed, big whoop! I just download into another body. I'm immortal, baby!"

	event.ResponseChannel <- response

	return nil
}
