package rps

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"reflect"
	"strings"
	"time"
)

type RPS struct {
	RpsPlaying string `json:"rps-playing"`
	Rps        string `json:"rps"`
	Name       string `json:"name"`
}

const rpsCacheTTL = time.Hour

type RpsBotGame struct {
	body            string
	sender          string
	target          string
	mm              *mmclient.MMClient
	settings        *settings.Settings
	ReplyChannel    *model.Channel
	ResponseChannel chan comms.Response
	Cache           cache.Cache
}

func Playing(player RPS) bool {
	return player.RpsPlaying != ""
}

func sameGame(player RPS, opponent RPS) bool {
	return player.RpsPlaying == "" || player.RpsPlaying == opponent.RpsPlaying
}

func differentUser(player RPS, opponent RPS) bool {
	return player.Name != opponent.Name
}

func FindApponent(event RpsBotGame, forPlayer RPS, chanId string) (RPS, error) {
	us, ok, err := users.GetUsers(event.Cache)
	rpsUs, ok, gPerr := getPlayers(us, chanId, event.Cache)
	var opponent RPS
	var found = false

	if us == nil {
		return RPS{}, fmt.Errorf("no opponent")
	}

	if ok {
		for _, u := range rpsUs {
			if Playing(u) && sameGame(forPlayer, u) && differentUser(forPlayer, u) {
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

func GetWinner(player RPS, opponent RPS) ([]RPS, bool) {
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

func GetPlayer(player users.User, chanId string, c cache.Cache) (RPS, error) {
	key := fmt.Sprintf("%s-%s-%s", "rps", player.Name, chanId)
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

func UpdateRps(playerRps RPS, chanId string, c cache.Cache) (RPS, error) {
	key := fmt.Sprintf("%s-%s-%s", "rps", playerRps.Name, chanId)
	p, _ := json.Marshal(playerRps)
	c.Put(key, p)
	setCacheTTL(c, key)
	return playerRps, nil
}

func DeleteRps(playerRps RPS, chanId string, c cache.Cache) {
	key := fmt.Sprintf("%s-%s-%s", "rps", playerRps.Name, chanId)
	c.Clean(key)
}
func DeleteGame(uuid string, c cache.Cache) {
	c.Clean(uuid)
}

// setCacheTTL sets an expiration when the underlying cache supports it.
func setCacheTTL(c cache.Cache, key string) {
	type expirer interface {
		Expire(string, time.Duration)
	}

	if ec, ok := c.(expirer); ok {
		ec.Expire(key, rpsCacheTTL)
	}
}

// SetGameTTL exposes TTL setting for game IDs stored outside this package.
func SetGameTTL(c cache.Cache, key string) {
	setCacheTTL(c, key)
}
