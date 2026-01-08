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
	var index int
	var err error

	if event.body == "" {
		arraySize := len(quotes)

		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		index = rand.Intn(arraySize)
	} else {
		index, err = strconv.Atoi(string(event.body[0]))
		if err != nil {
			return err
		}
	}
	response.Message = fmt.Sprintf(`%s`, quotes[index])
	response.Message = strings.Replace(response.Message, "{nick}", event.mm.BotUser.Username, -1)
	response.Message = strings.Replace(response.Message, "{0}", event.sender, -1)

	event.ResponseChannel <- response
	return nil
}

func (h BotCommandHelp) Quote(request BotCommand) (response HelpResponse) {
	response.Help = "Share a random quote or a specific quote by index. Usage: '!quote' or '!quote <number>'."
	response.Description = "Get a Bender quote"

	return response
}
