package games

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"reflect"
	"strings"
)

type (
	WHGameData struct {
		State   string               `json:"state"`
		Players []wavinghands.Wizard `json:"players"`
	}
	Game struct {
		gData   WHGameData
		Channel *model.Channel
	}
)

func NewWavingHands(event BotGame) (Game, error) {
	wHGameData, err := GetChannelGame(event.ReplyChannel.Id, event.cache)
	name := strings.TrimLeft(event.sender, "@")
	inGame := false

	if err != nil {
		w := wavinghands.Wizard{
			Right:       wavinghands.Hand{},
			Left:        wavinghands.Hand{},
			Name:        name,
			Living:      wavinghands.Living{},
			Curses:      "",
			Protections: "",
			Monsters:    wavinghands.Monster{},
		}
		wizards := append(wHGameData.Players, w)
		g := Game{gData: WHGameData{State: "starting", Players: wizards}, Channel: event.ReplyChannel}
		SetChannelGame(event.ReplyChannel.Id, g.gData, event.cache)
		return g, nil
	} else {
		g := Game{gData: wHGameData}
		wizards := wHGameData.Players
		switch g.gData.State {
		case "starting":
			for _, wR := range wizards {
				if inGame {
					break
				}

				inGame = wR.Name == name
			}
			if !inGame {
				w := wavinghands.Wizard{
					Right:       wavinghands.Hand{},
					Left:        wavinghands.Hand{},
					Name:        name,
					Living:      wavinghands.Living{},
					Curses:      "",
					Protections: "",
					Monsters:    wavinghands.Monster{},
				}
				wizards := append(wHGameData.Players, w)
				g := Game{gData: WHGameData{State: "starting", Players: wizards}, Channel: event.ReplyChannel}
				SetChannelGame(event.ReplyChannel.Id, g.gData, event.cache)
			} else {
				g := Game{gData: WHGameData{State: "starting", Players: wizards}, Channel: event.ReplyChannel}
				return g, fmt.Errorf("you're already in the game in %s channel", event.ReplyChannel.Name)
			}
			return g, nil
		case "playing":
		default:
			return Game{}, fmt.Errorf("game already in progress")
		}
	}

	return Game{}, fmt.Errorf("game not implemented")
}

func SetChannelGame(channelId string, g WHGameData, c cache.Cache) error {
	gS, _ := json.Marshal(g)
	c.Put(channelId, gS)
	return nil
}
func GetChannelGame(channelId string, c cache.Cache) (WHGameData, error) {
	g, ok, _ := c.Get(channelId)
	var wHGD WHGameData
	if ok {
		if reflect.TypeOf(g).String() != "[]uint8" {
			json.Unmarshal([]byte(g.(string)), &wHGD)
		} else {
			json.Unmarshal(g.([]byte), &wHGD)
		}
		return wHGD, nil
	}
	return WHGameData{}, fmt.Errorf("not found")
}

func (bg BotGame) Wh(event BotGame) (response Response, err error) {
	response.Type = "command"
	var directive string = event.body

	switch directive {
	default:
		g, err := NewWavingHands(event)
		if err != nil {
			response.Type = "dm"
			response.Message = err.Error()
			return response, nil
		}

		if len(g.gData.Players) > wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
			response.Message = fmt.Sprintf("/echo %s has joined waving hands. Would you like to start?", event.sender)
		} else if len(g.gData.Players) >= wavinghands.GetMaxTeams() {
			response.Message = fmt.Sprintf("/echo The waving hands game is full. Would you like to start?\n")
		} else if len(g.gData.Players) < wavinghands.GetMinTeams() {
			response.Message = fmt.Sprintf("/echo %s would like to play a game of Waving Hands.\n", event.sender)
		}
	}

	return response, nil
}
