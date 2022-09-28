package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strings"
)

func (bc BotCommand) S(event BotCommand) (response Response, err error) {
	var toReplace, withText string

	response.Type = "command"
	u, _ := users.GetUser(event.sender, event.cache)

	oldMessage := u.Message

	fmt.Sscanf(event.body, "!s/%s/%s/", &toReplace, &withText)

	newMessage := strings.Replace(oldMessage, toReplace, withText, -1)

	response.Message = fmt.Sprintf(`/echo "%s meant %s"`, event.sender, newMessage)

	return response, nil
}
