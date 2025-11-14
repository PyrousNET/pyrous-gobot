package commands

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

var rollRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (h BotCommandHelp) Roll(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Rolls two 6 sided dice for a random response to your query.\n e.g. !roll should I take a break?",
		Description: "Roll some dice!",
	}
}

func (bc BotCommand) Roll(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}

	dieSize := 5

	d1 := rollRand.Intn(dieSize) + 1
	d2 := rollRand.Intn(dieSize) + 1

	response.Message = fmt.Sprintf("%s rolled a %d and a %d for a total of %d", event.sender, d1, d2, d1+d2)

	event.ResponseChannel <- response
	return nil
}
