package games

import (
    "encoding/json"
    "fmt"
    "github.com/pyrousnet/pyrous-gobot/internal/cache"
    "reflect"
    "math/rand"
    "time"
    "strings"

    "github.com/google/uuid"
    "github.com/pyrousnet/pyrous-gobot/internal/users"
)

type Farkle struct {
    GameID          string      `json:"game-id"` //ID of this farkle game
    FarklePlaying   bool        `json:"farkle-playing"` //If the game has started
    FarkleMembers   []string    `json:"farkle-members"` //Members of the farkle game
    FarkleTotal     int         `json:"farkle-total"` //TODO: How to track scores
    FarkleTemp      int         `json:"farkle-temp"` //Track current rolling total
    DiceHolder      string      `json:"dice-holder"` //User currently holding the dice
    DiceThrown      []int       `json:"dice-thrown"` //Current dice list
    DiceToThrow     int         `json:"dice-to-throw"` //Number of dice which can be thrown
    Name            string      `json:"name"` //Name of the farkle user.
}

var DICE = 6;

func (h BotGameHelp) Farkle(request BotGame) (response HelpResponse) {
	response.Help = "Play Farkle in the current channel."
	response.Description = `Play a game of Farkle!\n
        Usage: $farklestart - Create a new game of Farkle.
        Usage: $farklejoin - Join a game of Farkle already created.
        Usage: $farkle - Throw your dice in the current game of Farkle (if they are yours to throw).
        Usage: $farkle n # - Attempt to rethrow # of dice.
        Usage: $farkle y - Keeps all the dice and ends your turn.` //TODO: $ should come from settings

	return response
}

func (bg BotGame) Farkle(event BotGame) (response Response, err error) {
    response.Type = "multi"
    var choice, channel string

    fmt.Sscanf(event.body, "%s %s", &channel, &choice) //TODO: Do we want to allow joining from other channels?
    foundChannel, cErr := event.mm.GetChannelByName(channel)
    if cErr == nil && foundChannel != nil {
        response.Channel = foundChannel.Id
    } else {
        foundChannel = event.ReplyChannel
    }
    playerUser, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache) //Gets the current user's details
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
        } else { //TODO: This is the logic to start a game.
            id, e := uuid.NewRandom()
            event.cache.Put(id.String(), event.ReplyChannel.Id)
            response.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps %s rock)##%s is looking for an opponent in RPS.", event.ReplyChannel.Name, event.sender)
            if e != nil {
                return response, e
            }
            player.RpsPlaying = id.String() //TODO: This is where the player is setup.
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

			deleteGame(player.RpsPlaying, event.cache)
			deleteRps(player, foundChannel.Id, event.cache)
			deleteRps(opponent, foundChannel.Id, event.cache)
			return response, err
		}

		updateRps(opponent, foundChannel.Id, event.cache)
	}

	updateRps(player, foundChannel.Id, event.cache)

	return response, err
}

func checkGame(players Farkle) (Farkle, error) {
}

func playing(player Farkle) bool {
	return player.GameID != ""
}

func getPlayer(player users.User, chanId string, c cache.Cache) (Farkle, error) {
	key := fmt.Sprintf("%s-%s-%s", "farkle", player.Name, chanId)
	var farkle Farkle
	r, ok, _ := c.Get(key)
	if ok {
		if reflect.TypeOf(r).String() != "[]uint8" {
			json.Unmarshal([]byte(r.(string)), &farkle)
		} else {
			json.Unmarshal(r.([]byte), &farkle)
		}
		return farkle, nil
	}
	return Farkle{Name: player.Name}, fmt.Errorf("not found")
}


//TODO: modify or remove the following functions.

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
	} else if strings.ToLower(player.Rps) == "rock" {
		if strings.ToLower(opponent.Rps) == "paper" {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == "scissors" {
			winners = append(winners, player)
			hasWinner = true
		}
	} else if strings.ToLower(player.Rps) == "paper" {
		if strings.ToLower(opponent.Rps) == "scissors" {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == "rock" {
			winners = append(winners, player)
			hasWinner = true
		}
	} else if strings.ToLower(player.Rps) == "scissors" {
		if strings.ToLower(opponent.Rps) == "rock" {
			winners = append(winners, opponent)
			hasWinner = true
		} else if strings.ToLower(opponent.Rps) == "paper" {
			winners = append(winners, player)
			hasWinner = true
		}
	}

	return winners, hasWinner
}

func getPlayers(pUsers []users.User, chanId string, c cache.Cache) ([]RPS, bool, error) {
	var rpsUsers []RPS
	for _, u := range pUsers {
		key := fmt.Sprintf("%s-%s-%s", "rps", u.Name, chanId)
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
	key := fmt.Sprintf("%s-%s-%s", "rps", playerRps.Name, chanId)
	p, _ := json.Marshal(playerRps)
	c.Put(key, p)
	return playerRps, nil
}

func deleteRps(playerRps RPS, chanId string, c cache.Cache) {
	key := fmt.Sprintf("%s-%s-%s", "rps", playerRps.Name, chanId)
	c.Clean(key)
}
func deleteGame(uuid string, c cache.Cache) {
	c.Clean(uuid)
}


