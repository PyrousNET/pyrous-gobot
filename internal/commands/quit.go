package commands

func (h BotCommandHelp) Quit(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Shuts down the bot",
		Description: "Shuts down the bot",
	}
}

func (bc BotCommand) Quit(event BotCommand) (response Response, err error) {
	response.Type = "shutdown"

	response.Message = "So, if my body gets killed, big whoop! I just download into another body. I'm immortal, baby!"

	return response, nil
}
