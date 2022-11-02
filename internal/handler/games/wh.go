package games

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands/spells"
	"reflect"
	"strings"
)

type (
	WHGameData struct {
		State   string               `json:"state"`
		Players []wavinghands.Wizard `json:"players"`
		Round   int                  `json:"round"`
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
			Right: wavinghands.Hand{},
			Left:  wavinghands.Hand{},
			Name:  name,
			Living: wavinghands.Living{
				HitPoints: 15,
			},
			Curses:      "",
			Protections: "",
			Monsters:    []wavinghands.Monster{},
		}
		wizards := append(wHGameData.Players, w)
		g := Game{gData: WHGameData{State: "starting", Players: wizards, Round: 0}, Channel: event.ReplyChannel}
		SetChannelGame(event.ReplyChannel.Id, g.gData, event.cache)
		return g, nil
	} else {
		g := Game{gData: wHGameData}
		wizards := wHGameData.Players
		switch g.gData.State {
		case "starting":
			inGame = isWizardInGame(wizards, inGame, name)
			if !inGame {
				w := wavinghands.Wizard{
					Right: wavinghands.Hand{},
					Left:  wavinghands.Hand{},
					Name:  name,
					Living: wavinghands.Living{
						HitPoints: 15,
					},
					Curses:      "",
					Protections: "",
					Monsters:    []wavinghands.Monster{},
				}
				wizards := append(wHGameData.Players, w)
				g = Game{gData: WHGameData{State: "starting", Players: wizards, Round: 0}, Channel: event.ReplyChannel}
				SetChannelGame(event.ReplyChannel.Id, g.gData, event.cache)
			} else {
				g = Game{gData: WHGameData{State: "starting", Players: wizards, Round: 0}, Channel: event.ReplyChannel}
				return g, fmt.Errorf("you're already in the game in %s channel", event.ReplyChannel.Name)
			}
			return g, nil
		default:
			return Game{}, fmt.Errorf("game already in progress")
		}
	}

	return Game{}, fmt.Errorf("game not implemented")
}

func StartWavingHands(event BotGame) (Game, error) {
	wHGameData, err := GetChannelGame(event.ReplyChannel.Id, event.cache)
	g := Game{gData: wHGameData, Channel: event.ReplyChannel}
	name := strings.TrimLeft(event.sender, "@")
	inGame := false

	if err != nil {
		return Game{}, err
	}
	wizards := wHGameData.Players
	switch g.gData.State {
	case "starting":
		inGame = isWizardInGame(wizards, inGame, name)
		if !inGame {
			return Game{}, fmt.Errorf("player not active in game, cannot start")
		}
		if len(g.gData.Players) > wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
			g.gData.State = "playing"
		} else if len(g.gData.Players) < wavinghands.GetMinTeams() {
			return Game{}, fmt.Errorf("not enough players to start the game")
		}
		SetChannelGame(event.ReplyChannel.Id, g.gData, event.cache)
	default:
		return Game{}, fmt.Errorf("cannot start game at this time")
	}

	return g, nil
}

func PromptForGestures(g Game, event BotGame) error {

	return nil
}

func isWizardInGame(wizards []wavinghands.Wizard, inGame bool, name string) bool {
	for _, wR := range wizards {
		if inGame {
			break
		}

		inGame = wR.Name == name
	}
	return inGame
}

func SetChannelGame(channelId string, g WHGameData, c cache.Cache) error {
	gS, _ := json.Marshal(g)
	c.Put(channelId, gS)
	return nil
}
func ClearGame(channelId string, c cache.Cache) error {
	c.Clean(channelId)
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

	if event.body == "" {
		r, err, done := handleEmptyBody(event, response)
		if done {
			return r, err
		}
	} else {
		r, err, done := handleGameWithDirective(event, response, err)
		if done {
			return r, err
		}
	}

	return response, nil
}

func handleGameWithDirective(event BotGame, response Response, err error) (Response, error, bool) {
	parts := strings.Split(event.body, " ")
	directive := parts[0]
	switch directive {
	case "help":
		response.Channel = event.ReplyChannel.Id
		response.Message = wavinghands.GetHelpSpell(parts[1])
		return response, nil, true
	case "help-spells":
		response.Channel = event.ReplyChannel.Id
		response.Message = wavinghands.GetHelpSpells()
		return response, nil, true
	case "start":
		g, err := StartWavingHands(event)
		if err != nil {
			response.Type = "dm"
			response.Message = err.Error()
			return response, nil, true
		}
		response.Type = "multi"
		messages := []string{}
		for _, w := range g.gData.Players {
			messages = append(messages, fmt.Sprintf("%s;;Please supply 2 gestures for %s (f, p, s, w, d, c {requires both hands}, stab, nothing)", w.Name, g.Channel.Name))
		}
		messages = append(messages, "This game of Waving Hands has begun!")
		response.Message = strings.Join(messages, "##")
	default:
		var channelName, rGesture, lGesture, target string
		var wHGameData WHGameData
		var t *wavinghands.Living
		name := strings.TrimLeft(event.sender, "@")
		if err != nil {
			return response, err, true
		}
		fmt.Sscanf(event.body, "%s %s %s %s", &channelName, &rGesture, &lGesture, &target)
		if channelName != "" {
			channel, err := event.mm.GetChannelByName(channelName)
			if err != nil {
				return response, err, true
			}
			response.Channel = channel.Id
			wHGameData, err = GetChannelGame(channel.Id, event.cache)
		} else {
			return response, fmt.Errorf("no channel name included"), true
		}

		g := Game{gData: wHGameData, Channel: event.ReplyChannel}
		p, err := GetCurrentPlayer(g, name)

		if target != "" {
			t, err = FindTarget(g, target)
		}

		if err != nil {
			return response, err, true
		} else {
			if rGesture == "stab" {
				rGesture = "1"
			}
			if lGesture == "stab" {
				lGesture = "1"
			}
			if rGesture == "nothing" {
				rGesture = "0"
			}
			if rGesture == "stab" {
				lGesture = "0"
			}
			rightGestures := append(p.Right.Get(), rGesture[0])
			leftGestures := append(p.Left.Get(), lGesture[0])

			p.Right.Set(string(rightGestures))
			p.Left.Set(string(leftGestures))
		}

		// Check All Players for gestures
		hasAllMoves := CheckAllPlayers(g)
		if hasAllMoves {
			sr, err := spells.GetSurrenderSpell(wavinghands.GetSpell("Surrender"))
			if err != nil {
				return response, err, true
			}
			surrenderString, err := sr.Cast(p, t)
			if err == nil && surrenderString != "" {
				response.Type = "multi"
				response.Message = surrenderString + "##"
			} else if err != nil {
				return response, err, true
			}
			// Run Protection Spells

			cHW, err := spells.GetCureHeavyWoundsSpell(wavinghands.GetSpell("Cure Heavy Wounds"))
			if err != nil {
				return response, err, true
			}
			chwResult, err := cHW.Cast(p, t)
			if err == nil && chwResult != "" {
				response.Message += chwResult + "##"
			} else if err != nil {
				return response, err, true
			}

			// Run Damage Spells
			CHW, err := spells.GetCauseHeavyWoundsSpell(wavinghands.GetSpell("Cause Heavy Wounds"))
			if err != nil {
				return response, err, true
			}
			chwResult, err = CHW.Cast(p, t)
			if err == nil && chwResult != "" {
				response.Message += chwResult + "##"
			} else if err != nil {
				return response, err, true
			}

			// Run Summon Spells
			// Clear Player Gestures
			ClearGestures(g)
			g.gData.Round += 1
		}

		winner, err := getWHWinner(g)
		if err == nil {
			response.Message += fmt.Sprintf("%s has won the game of waving hands.", winner.Name)

			ClearGame(g.Channel.Id, event.cache)
		} else {
			SetChannelGame(g.Channel.Id, g.gData, event.cache)
		}
	}
	return Response{}, nil, false
}

func handleEmptyBody(event BotGame, response Response) (Response, error, bool) {
	g, err := NewWavingHands(event)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		return response, nil, true
	}

	if len(g.gData.Players) >= wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo %s has joined waving hands. Would you like to start?", event.sender)
	} else if len(g.gData.Players) >= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo The waving hands game is full. Would you like to start?\n")
	} else if len(g.gData.Players) < wavinghands.GetMinTeams() {
		response.Message = fmt.Sprintf("/echo %s would like to play a game of Waving Hands.\n", event.sender)
	}
	return Response{}, nil, false
}

func getWHWinner(g Game) (wavinghands.Wizard, error) {
	var winner wavinghands.Wizard
	var found = false
	for _, w := range g.gData.Players {
		if !found && w.Living.HitPoints > 0 {
			winner = w
			found = true
		} else if w.Living.HitPoints > 0 {
			return wavinghands.Wizard{}, fmt.Errorf("no winner yet")
		}
	}

	if found {
		return winner, nil
	} else {
		return wavinghands.Wizard{}, fmt.Errorf("no winner yet")
	}
}

func FindTarget(g Game, selector string) (*wavinghands.Living, error) {
	var name, monster string
	var wizard *wavinghands.Wizard
	parts := strings.Split(selector, ":")
	if len(parts) > 1 {
		name = parts[0]
		monster = parts[1]
	} else {
		name = selector
	}
	for i, w := range g.gData.Players {
		if w.Name == name {
			wizard = &g.gData.Players[i]
		}

		if monster == "" {
			wizard.Living.Selector = selector
			return &wizard.Living, nil
		}
	}

	if monster != "" {
		for i, m := range wizard.Monsters {
			if m.Type == monster {
				wizard.Living.Selector = selector
				return &wizard.Monsters[i].Living, nil
			}
		}
	}

	return &wavinghands.Living{}, fmt.Errorf("not found")
}

func GetCurrentPlayer(g Game, name string) (*wavinghands.Wizard, error) {
	for i, w := range g.gData.Players {
		if w.Name == name {
			return &g.gData.Players[i], nil
		}
	}

	return &wavinghands.Wizard{}, fmt.Errorf("not found")
}

func ClearGestures(g Game) {
	for _, w := range g.gData.Players {
		w.Right.Set("")
		w.Left.Set("")
	}
}

func CheckAllPlayers(g Game) bool {
	for _, w := range g.gData.Players {
		if w.Living.HitPoints <= 0 { // Dead wizards have no gestures
			continue
		}
		if len(w.Right.Get()) <= g.gData.Round {
			return false
		}
		if len(w.Left.Get()) <= g.gData.Round {
			return false
		}
	}
	return true
}