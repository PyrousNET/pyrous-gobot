package commands

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Quote(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}

	quotes := event.settings.GetQuotes()
	if len(quotes) == 0 {
		response.Message = "No quotes are configured."
		event.ResponseChannel <- response
		return nil
	}
	var index int
	var err error

	if event.body == "" {
		arraySize := len(quotes)

		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		index = rand.Intn(arraySize)
	} else {
		args := strings.Fields(event.body)
		index, err = strconv.Atoi(args[0])
		if err != nil {
			response.Message = "Invalid quote index. Use !quote or !quote <number>."
			event.ResponseChannel <- response
			return nil
		}
	}
	if index < 0 || index >= len(quotes) {
		response.Message = fmt.Sprintf("Quote %d not found.", index)
		event.ResponseChannel <- response
		return nil
	}

	response.Message = fmt.Sprintf(`%s`, quotes[index])
	if event.mm != nil && event.mm.BotUser != nil {
		response.Message = strings.Replace(response.Message, "{nick}", event.mm.BotUser.Username, -1)
	}
	response.Message = strings.Replace(response.Message, "{0}", event.sender, -1)

	event.ResponseChannel <- response
	return nil
}
