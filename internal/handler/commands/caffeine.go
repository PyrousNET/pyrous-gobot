package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"strconv"
)

func (c BotCommand) Caffeine(event BotCommand) error {
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
	}
	response.Type = "command"

	if event.body != "" {
		numShots, err := strconv.Atoi(string(event.body[0]))
		if err == nil {
			fmt.Printf("%+v\n", numShots)
			response.Message = fmt.Sprintf("/me walks over to %s and gives them %d shots of caffeine straight into the blood stream.", event.sender, numShots)
		}
	} else {
		response.Message = fmt.Sprintf("/me walks over to %s and gives them a shot of caffeine straight into the blood stream.", event.sender)
	}

	event.ResponseChannel <- response

	return nil
}
