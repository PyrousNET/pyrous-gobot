package games

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"strings"

	"github.com/google/uuid"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const ROCK = "rock"
const PAPER = "paper"
const SCISSORS = "scissors"

type RPS struct {
	RpsPlaying string `json:"rps-playing"`
	Rps        string `json:"rps"`
	Name       string `json:"name"`
}

func (bg BotGame) Rps(event BotGame) (response Response, err error) {
	response.Type = "multi"
	playerUser, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil {
		return Response{}, err
	}
	player, err := getPlayer(playerUser)
	opponent, oErr := findApponent(event, player)
	if !playing(player) {
		if oErr == nil && playing(opponent) {
			channelId, ok, _ := event.cache.Get(opponent.RpsPlaying)
			if ok && event.ReplyChannel != nil && channelId == event.ReplyChannel.Id {
				player.RpsPlaying = opponent.RpsPlaying
				response.Type = "dm"
				response.Message = "Would you like to throw Rock, Paper or Scissors (Usage: $rps rock)"
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

		updateRps(opponent, event.cache)
	}

	updateRps(player, event.cache)

	return response, err
}

func playing(player RPS) bool {
	return player.RpsPlaying != ""
}

func sameGame(player RPS, opponent RPS) bool {
	return player.RpsPlaying == "" || player.RpsPlaying == opponent.RpsPlaying
}

func differentUser(player RPS, opponent RPS) bool {
	return player.Name != opponent.Name
}

func findApponent(event BotGame, forPlayer RPS) (RPS, error) {
	us, ok, err := users.GetUsers(event.cache)
	rpsUs, ok, err := getPlayers(us, event.cache)
	var opponent RPS
	var found = false

	if us == nil {
		return RPS{}, fmt.Errorf("no opponent")
	}

	if ok {
		for _, u := range rpsUs {
			if playing(u) && sameGame(forPlayer, u) && differentUser(forPlayer, u) {
				opponent = u
				found = true
			}
		}
	} else {
		return RPS{}, err
	}

	if !found {
		err = fmt.Errorf("no opponent")
	}
	return opponent, err
}

func getWinner(player RPS, opponent RPS) ([]RPS, bool) {
	var hasWinner bool = false
	var winners []RPS

	if player.Rps == "" {
		return []RPS{}, false
	}

	if opponent.Rps == "" {
		return []RPS{}, false
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

func getPlayer(player users.User) (RPS, error) {
	// TODO
	return RPS{}, nil
}

func getPlayers(pUsers []users.User, c cache.Cache) ([]RPS, bool, error) {
	// TODO
	return []RPS{}, true, nil
}

func updateRps(playerRps RPS, c cache.Cache) (RPS, error) {
	// TODO
	return RPS{}, nil
}
