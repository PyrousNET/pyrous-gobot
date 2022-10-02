package games

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"strings"
)

const ROCK = "rock"
const PAPER = "paper"
const SCISSORS = "scissors"

func (bg BotGame) Rps(event BotGame) (response Response, err error) {
	response.Type = "multi"
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	opponent, oErr := findApponent(event, player)
	if !playing(player) {
		if oErr == nil && playing(opponent) {
			channelId, ok, _ := event.cache.Get(opponent.RpsPlaying)
			if ok && event.ReplyChannel != nil && channelId == event.ReplyChannel.Id {
				player.RpsPlaying = opponent.RpsPlaying
			}
		} else {
			id, e := uuid.NewRandom()
			event.cache.Put(id.String(), event.ReplyChannel.Id)
			response.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps rock)##%s is looking for an opponent in RPS.", event.sender)
			if e != nil {
				return response, e
			}
			player.RpsPlaying = id.String()
		}
	}

	if event.body != "" {
		switch strings.ToLower(event.body) {
		case "rock", "paper", "scissors":
			player.Rps = strings.ToLower(event.body)
			response.Type = "dm"
			response.Message = fmt.Sprintf("I have you down for: %s", strings.Title(strings.ToLower(event.body)))
		default:
			response.Type = "dm"
			response.Message = fmt.Sprintf(`Uh, %s isn't an option. Try rock, paper or scissors'`, event.body)
		}
	}

	if oErr == nil {
		winners, hasWinner := getWinner(player, opponent)
		if hasWinner {
			channelId, ok, _ := event.cache.Get(player.RpsPlaying)
			response.Type = "command"
			if ok {
				response.Channel = channelId.(string)
				if len(winners) > 1 {
					response.Message = fmt.Sprintf("/echo The RPS game between %s and %s ended in a draw.", player.Name, opponent.Name)
				} else {
					response.Message = fmt.Sprintf("/echo The RPS game between %s and %s ended with %s winning.", player.Name, opponent.Name, winners[0].Name)
				}
			}

			player.Rps = ""
			player.RpsPlaying = ""
			opponent.Rps = ""
			opponent.RpsPlaying = ""
		}

		users.UpdateUser(opponent, event.cache)
	}

	users.UpdateUser(player, event.cache)

	return response, err
}

func playing(player users.User) bool {
	return player.RpsPlaying != ""
}

func sameGame(player users.User, opponent users.User) bool {
	return player.RpsPlaying == "" || player.RpsPlaying == opponent.RpsPlaying
}

func differentUser(player users.User, opponent users.User) bool {
	return player.Name != opponent.Name
}

func findApponent(event BotGame, forPlayer users.User) (users.User, error) {
	us, ok, err := users.GetUsers(event.cache)
	var opponent users.User
	var found = false

	if us == nil {
		return users.User{}, fmt.Errorf("no opponent")
	}

	if ok {
		for _, u := range us {
			if playing(u) && sameGame(forPlayer, u) && differentUser(forPlayer, u) {
				opponent = u
				found = true
			}
		}
	} else {
		return users.User{}, err
	}

	if !found {
		err = fmt.Errorf("no opponent")
	}
	return opponent, err
}

func getWinner(player users.User, opponent users.User) ([]users.User, bool) {
	var hasWinner bool = false
	var winners []users.User

	if player.Rps == "" {
		return []users.User{}, false
	}

	if opponent.Rps == "" {
		return []users.User{}, false
	}

	if player.Rps == opponent.Rps {
		winners = append(winners, player)
		winners = append(winners, opponent)
		hasWinner = true
	} else if strings.ToLower(player.Rps) == ROCK {
		if strings.ToLower(opponent.Rps) == PAPER {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == SCISSORS {
			winners = append(winners, player)
			hasWinner = true
		}
	} else if strings.ToLower(player.Rps) == PAPER {
		if strings.ToLower(opponent.Rps) == SCISSORS {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == ROCK {
			winners = append(winners, player)
			hasWinner = true
		}
	} else if strings.ToLower(player.Rps) == SCISSORS {
		if strings.ToLower(opponent.Rps) == ROCK {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == PAPER {
			winners = append(winners, player)
			hasWinner = true
		}
	}

	return winners, hasWinner
}
