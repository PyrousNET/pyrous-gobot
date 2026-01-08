package commands

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bc BotCommand) Insult(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}

	insults := event.settings.GetInsults()
	var index int

	if event.body == "" {
		response.Type = "dm"
		response.Message = "You must tell me who to insult"

		event.ResponseChannel <- response
		return nil
	}
	_, ok, _ := users.GetUser(event.body, event.cache)
	if ok {
		arraySize := len(insults)

		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		index = rand.Intn(arraySize)
		response.Message = fmt.Sprintf(`%s`, insults[index])
		response.Message = strings.Replace(response.Message, "{nick}", event.mm.BotUser.Username, -1)
		response.Message = strings.Replace(response.Message, "{0}", event.body, -1)
	} else {
		response.Type = "dm"
		response.Message = fmt.Sprintf("Who's %s?", event.body)
	}

	event.ResponseChannel <- response
	return nil
}
