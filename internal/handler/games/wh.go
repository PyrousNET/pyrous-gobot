package games

import "fmt"

func (bg BotGame) Wh(event BotGame) (response Response, err error) {
	response.Type = "command"
	var directive string = event.body

	switch directive {
	case "join":
		response.Message = fmt.Sprintf("/echo %s would like to play a game of Waving Hands.", event.sender)
	default:
		response.Type = "dm"
		response.Message = fmt.Sprintf("Would you like the join a game of Waving Hands?")
	}

	return response, nil
}
