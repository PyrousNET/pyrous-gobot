package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (h BotCommandHelp) S(request BotCommand) (response HelpResponse) {
	response.Help = "Have Bender replace a typo in the last thing you just said."

	response.Description = "Cause the bot to replace the string for you. Usage: '!s /{old}/new/'"

	return response
}

func (bc BotCommand) S(event BotCommand) error {
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
	}
	var toReplace, withText string

	response.Type = "command"
	u, ok, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)

	if err != nil {
		return err
	}

	var oldMessage string
	if ok {
		oldMessage = u.Message
	}

	parts := strings.Split(event.body, "/")
	if len(parts) < 3 {
		response.Type = "dm"
		response.Message = "Incorrect string replace format. Try !help s"

		return nil
	}
	toReplace = parts[1]
	withText = parts[2]

	newMessage := strings.Replace(oldMessage, toReplace, withText, -1)

	response.Message = fmt.Sprintf(`/echo %s meant: "%s"`, event.sender, newMessage)

	event.ResponseChannel <- response

	return nil
}
