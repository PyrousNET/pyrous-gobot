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

var rollRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func (h BotCommandHelp) Roll(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Rolls custom dice using '!roll NdM reason' (e.g. !roll 3d6 attack). If no NdM spec is provided, defaults to two 6-sided dice for a decision roll.",
		Description: "Roll dice with optional NdM spec",
	}
}

func (bc BotCommand) Roll(event BotCommand) error {
	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}

	count, sides, reason, err := parseRollSpec(event.body)
	if err != nil {
		response.Message = err.Error()
		event.ResponseChannel <- response
		return nil
	}

	formattedSpec := fmt.Sprintf("%dd%d", count, sides)
	rolls := make([]int, count)
	total := 0

	for i := 0; i < count; i++ {
		rolls[i] = rollRand.Intn(sides) + 1
		total += rolls[i]
	}

	var message string
	if count == 2 && sides == 6 {
		message = fmt.Sprintf("%s rolled a %d and a %d for a total of %d", event.sender, rolls[0], rolls[1], total)
	} else if count == 1 {
		message = fmt.Sprintf("%s rolled %s and got %d", event.sender, formattedSpec, total)
	} else {
		parts := make([]string, len(rolls))
		for i, roll := range rolls {
			parts[i] = strconv.Itoa(roll)
		}
		message = fmt.Sprintf("%s rolled %s (%s) for a total of %d", event.sender, formattedSpec, strings.Join(parts, " + "), total)
	}

	if reason != "" {
		message = fmt.Sprintf("%s - %s", message, reason)
	}

	response.Message = message
	event.ResponseChannel <- response
	return nil
}
