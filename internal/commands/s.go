package commands

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strings"
)

func (bc BotCommand) S(event BotCommand) (response Response, err error) {
	var toReplace, withText string

	response.Type = "command"
	u, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)

	if err != nil {
		return Response{}, err
	}

	oldMessage := u.Message

	parts := strings.Split(event.body, "/")
	if len(parts) < 3 {
		return Response{}, fmt.Errorf("%s", "Incorrect string replace format.")
	}
	toReplace = parts[1]
	withText = parts[2]

	newMessage := strings.Replace(oldMessage, toReplace, withText, -1)

	response.Message = fmt.Sprintf(`/echo %s meant: "%s"`, event.sender, newMessage)

	return response, nil
}
