package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"math/rand"
	"strings"
	"time"
)

func (h BotCommandHelp) Praise(request BotCommand) (response HelpResponse) {
	response.Help = "Give a Bender specific praise from a random list."

	response.Description = "Cause the bot to praise someone. Usage: '!praise {target}'"

	return response
}

func (bc BotCommand) Praise(event BotCommand) (response Response, err error) {
	praises := event.settings.GetPraises()
	response.Type = "post"
	var index int

	if event.body == "" {
		response.Type = "dm"
		response.Message = "You must tell me who to praise"

		return response, nil
	}
	u, err := users.HasUser(event.body, event.cache)
	if u {
		arraySize := len(praises)

		rand := rand.New(rand.NewSource(time.Now().UnixNano()))
		index = rand.Intn(arraySize)
		response.Message = fmt.Sprintf(`%s`, praises[index])
		response.Message = strings.Replace(response.Message, "{nick}", event.mm.BotUser.Username, -1)
		response.Message = strings.Replace(response.Message, "{0}", event.body, -1)
	} else {
		response.Type = "dm"
		response.Message = fmt.Sprintf(`Who's ` + event.body + `?`)
	}

	return response, nil
}
