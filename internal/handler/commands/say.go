package commands

import (
	"fmt"
)

func (h BotCommandHelp) Say(request BotCommand) (response HelpResponse) {
    response.Help = "Give Bender a line of text to say in a channel."

    response.Description = "Cause the bot to say something in a channel. \n" +
    "Usage: '!say in {channel} {text}' or '!say {text}' for same channel."

    return response
}

func (bc BotCommand) Say(event BotCommand) (response Response, err error) {
	response.Type = "command"
	response.Message = fmt.Sprintf(`/echo "%s" 1`, event.body)

	return response, nil
}
