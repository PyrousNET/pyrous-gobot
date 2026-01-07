package commands

import (
	"fmt"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (h BotCommandHelp) React(request BotCommand) (response HelpResponse) {
	reactions := request.settings.GetReactions()
	var m string
	for i, r := range reactions {
		m += i + " - " + r.Description + "\n"
	}
	response.Help = m

	response.Description = "Curated reactions. Mostly gifs. Usage: '!react {reaction}'"

	return response
}

func (bc BotCommand) React(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
	}

	reactions := event.settings.GetReactions()
	key := strings.ToLower(strings.Join(strings.Fields(event.body), " "))
	if r, ok := reactions[key]; ok {
		response.Type = "command"
		response.Message = fmt.Sprintf(`/echo "![%s](%s)" 1`, r.Description, r.Url)
	} else {
		response.Type = "post"
		err := fmt.Errorf("Response key '%s' not found.", event.body)
		response.Message = fmt.Sprintf("%s", err)
	}

	event.ResponseChannel <- response
	return nil
}
