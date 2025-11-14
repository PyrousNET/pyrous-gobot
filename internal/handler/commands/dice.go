package commands

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

var (
	diceRand           = rand.New(rand.NewSource(time.Now().UnixNano()))
	diceCommandPattern = regexp.MustCompile(`(?i)^(\d*)d(\d+)$`)
)

const maxDice = 50

func (h BotCommandHelp) Dice(request BotCommand) (response HelpResponse) {
	return HelpResponse{
		Help:        "Rolls custom dice using shorthand like '!2d6 attack' or '!1d20'. First number is how many dice, second is the sides.",
		Description: "Roll arbitrary dice",
	}
}

func (bc BotCommand) Dice(event BotCommand) error {
	diceSpec := strings.ToLower(event.target)
	if diceSpec == "" {
		diceSpec = "1d20"
	}

	matches := diceCommandPattern.FindStringSubmatch(diceSpec)
	if len(matches) == 0 {
		return fmt.Errorf("invalid dice format '%s'. Use something like !2d6", diceSpec)
	}

	count := 1
	if matches[1] != "" {
		var err error
		count, err = strconv.Atoi(matches[1])
		if err != nil || count <= 0 {
			return fmt.Errorf("invalid dice count '%s'", matches[1])
		}
	}

	sides, err := strconv.Atoi(matches[2])
	if err != nil || sides <= 0 {
		return fmt.Errorf("invalid dice sides '%s'", matches[2])
	}

	if count > maxDice {
		return fmt.Errorf("that's too many dice! Please roll %d or fewer.", maxDice)
	}

	u, _, _ := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
	}

	formattedSpec := fmt.Sprintf("%dd%d", count, sides)
	rolls := make([]int, count)
	total := 0

	for i := 0; i < count; i++ {
		rolls[i] = diceRand.Intn(sides) + 1
		total += rolls[i]
	}

	reason := strings.TrimSpace(event.body)
	var message string

	if count == 1 {
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
