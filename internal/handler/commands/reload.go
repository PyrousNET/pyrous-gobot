package commands

func (h BotCommandHelp) Reload(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Causes the bot to shutdown, pull any changes from git, and restart",
		Description: "Reloads the bot",
	}
}

func (bc BotCommand) Reload(event BotCommand) (response Response, err error) {
	response.Type = "shutdown"

	response.Message = "So, if my body gets killed, big whoop! I just download into another body. I'm immortal, baby!"

	return response, nil
}
