package games

import "fmt"

func (bg BotGame) Farkle(event BotGame) (response Response, err error) {
	response.Type = "command"
	response.Message = fmt.Sprintf("/echo %s would like to play a game of Farkle.", event.sender)

	// TODO - for kr0w?

	return response, nil
}


