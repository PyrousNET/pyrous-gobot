package games

import (
	"encoding/json"
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const ROCK = "rock"
const PAPER = "paper"
const SCISSORS = "scissors"
const PREFIX = "rps-"

type RPS struct {
	RpsPlaying string `json:"rps-playing"`
	Rps        string `json:"rps"`
	Name       string `json:"name"`
}

func (bg BotGame) Rps(event BotGame) (response Response, err error) {
	response.Type = "multi"
	var choice, channel string
	fmt.Sscanf(event.body, "%s %s", &channel, &choice)
	foundChannel, cErr := event.mm.GetChannelByName(channel)
	if cErr == nil && foundChannel != nil {
		response.Channel = foundChannel.Id
	} else {
		foundChannel = event.ReplyChannel
	}
	playerUser, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil {
		return Response{}, err
	}
	player, perr := getPlayer(playerUser, foundChannel.Id, event.cache)
	opponent, oErr := findApponent(event, player, foundChannel.Id)
	if perr != nil || !playing(player) {
		if oErr == nil && playing(opponent) {
			channelId, ok, _ := event.cache.Get(opponent.RpsPlaying)
			if ok && event.ReplyChannel != nil && channelId == foundChannel.Id {
				player.RpsPlaying = opponent.RpsPlaying
				response.Type = "dm"
				response.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps %s rock)", event.ReplyChannel.Name)
			}
		} else {
			id, e := uuid.NewRandom()
			event.cache.Put(id.String(), event.ReplyChannel.Id)
			response.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps %s rock)##%s is looking for an opponent in RPS.", event.ReplyChannel.Name, event.sender)
			if e != nil {
				return response, e
			}
			player.RpsPlaying = id.String()
		}
	}

	if event.body != "" {
		switch strings.ToLower(choice) {
		case "rock", "paper", "scissors":
			player.Rps = strings.ToLower(choice)
			response.Type = "dm"
			response.Message = fmt.Sprintf("I have you down for: %s", strings.Title(strings.ToLower(choice)))
		default:
			response.Type = "dm"
			response.Message = fmt.Sprintf(`Uh, %s isn't an option. Try rock, paper or scissors'`, choice)
		}
	}

	if oErr == nil && opponent.Name != "" {
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

		updateRps(opponent, event.ReplyChannel.Id, event.cache)
	}

	updateRps(player, event.ReplyChannel.Id, event.cache)

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

func findApponent(event BotGame, forPlayer RPS, chanId string) (RPS, error) {
	us, ok, err := users.GetUsers(event.cache)
	rpsUs, ok, gPerr := getPlayers(us, chanId, event.cache)
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
		return RPS{}, gPerr
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

func getPlayer(player users.User, chanId string, c cache.Cache) (RPS, error) {
	var key string = PREFIX + player.Name + "-" + chanId
	var rps RPS
	r, ok, _ := c.Get(key)
	if ok {
		if reflect.TypeOf(r).String() != "[]uint8" {
			json.Unmarshal([]byte(r.(string)), &rps)
		} else {
			json.Unmarshal(r.([]byte), &rps)
		}
		return rps, nil
	}
	return RPS{Name: player.Name}, fmt.Errorf("not found")
}

func getPlayers(pUsers []users.User, chanId string, c cache.Cache) ([]RPS, bool, error) {
	var rpsUsers []RPS
	for _, u := range pUsers {
		var key string = PREFIX + u.Name + "-" + chanId
		r, ok, _ := c.Get(key)
		var rps RPS
		if ok {
			if reflect.TypeOf(r).String() != "[]uint8" {
				json.Unmarshal([]byte(r.(string)), &rps)
			} else {
				json.Unmarshal(r.([]byte), &rps)
			}

			rpsUsers = append(rpsUsers, rps)
		}
	}
	if len(rpsUsers) == 0 {
		return []RPS{}, false, nil
	}
	return rpsUsers, true, nil
}

func updateRps(playerRps RPS, chanId string, c cache.Cache) (RPS, error) {
	var key string = PREFIX + playerRps.Name + "-" + chanId
	p, _ := json.Marshal(playerRps)
	c.Put(key, p)
	return playerRps, nil
}
