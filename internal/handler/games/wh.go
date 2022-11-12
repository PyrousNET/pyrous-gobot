package games

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands/spells"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"golang.org/x/exp/slices"
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
	wHGameData, err := GetChannelGame(event.ReplyChannel.Id, event.Cache)
	name := strings.TrimLeft(event.sender, "@")
	if name == "" {
		return Game{}, fmt.Errorf("player is missing a name")
	}
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
		SetChannelGame(event.ReplyChannel.Id, g.gData, event.Cache)
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
				SetChannelGame(event.ReplyChannel.Id, g.gData, event.Cache)
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
	wHGameData, err := GetChannelGame(event.ReplyChannel.Id, event.Cache)
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
		SetChannelGame(event.ReplyChannel.Id, g.gData, event.Cache)
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
	c.Put(wavinghands.PREFIX+channelId, gS)
	return nil
}
func ClearGame(channelId string, c cache.Cache) error {
	c.Clean(wavinghands.PREFIX + channelId)
	return nil
}
func GetChannelGame(channelId string, c cache.Cache) (WHGameData, error) {
	g, ok, _ := c.Get(wavinghands.PREFIX + channelId)
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

func (bg BotGame) Wh(event BotGame) error {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
		Quit:           nil,
	}

	if event.body == "" {
		err, done := handleEmptyBody(event)
		if done {
			event.ResponseChannel <- response
			return err
		}
	} else {
		err, done := handleGameWithDirective(event, err)
		if done {
			return err
		}
	}

	return nil
}

func handleGameWithDirective(event BotGame, err error) (error, bool) {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
		Quit:           nil,
	}
	parts := strings.Split(event.body, " ")
	directive := parts[0]
	switch directive {
	case "help":
		response.Message = wavinghands.GetHelpSpell(parts[1])
		response.Type = "dm"
		event.ResponseChannel <- response
		return nil, true
	case "help-spells":
		response.Message = wavinghands.GetHelpSpells()
		response.Type = "dm"
		event.ResponseChannel <- response
		return nil, true
	case "start":
		g, err := StartWavingHands(event)
		if err != nil {
			response.Type = "dm"
			response.Message = err.Error()
			event.ResponseChannel <- response
			return nil, true
		}
		for _, w := range g.gData.Players {
			response.Type = "dm"
			player, _, err := users.GetUser(w.Name, event.Cache)
			if err != nil {
				response.Type = "dm"
				response.Message = err.Error()
				event.ResponseChannel <- response
				return nil, true
			}
			response.UserId = player.Id
			response.Message = fmt.Sprintf(
				"Please supply 2 gestures for %s (f, p, s, w, d, c {requires both hands}, stab, nothing)",
				g.Channel.Name)
			event.ResponseChannel <- response
		}

		response.Type = "command"
		response.Message = "/echo This game of Waving Hands has begun!"
		event.ResponseChannel <- response
	default:
		var channelName, rGesture, lGesture, target string
		var channel *model.Channel
		var wHGameData WHGameData
		var t *wavinghands.Living
		name := strings.TrimLeft(event.sender, "@")
		if err != nil {
			return err, true
		}
		fmt.Sscanf(event.body, "%s %s %s %s", &channelName, &rGesture, &lGesture, &target)
		if channelName != "" {
			c, err := event.mm.GetChannelByName(channelName)
			if err != nil {
				return err, true
			}
			channel = c
			response.ReplyChannelId = channel.Id
			wHGameData, err = GetChannelGame(channel.Id, event.Cache)
		} else {
			return fmt.Errorf("no channel name included"), true
		}

		g := Game{gData: wHGameData, Channel: channel}
		p, err := GetCurrentPlayer(g, name)

		if err != nil {
			return err, true
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
			p.SetTarget(target)
		}
		// Completed setting gestures for the current player

		// Check All Players for gestures
		hasAllMoves := CheckAllPlayers(g)
		if hasAllMoves {
			for i, p := range g.gData.Players {
				rG := p.Right.GetAt(len(p.Right.Sequence) - 1)
				lG := p.Left.GetAt(len(p.Left.Sequence) - 1)
				if p.Target != "" {
					t, err = FindTarget(g, p.Target)
				}
				announceGestures(&p, event.ResponseChannel, response, string(rG), string(lG), p.GetTarget())
				sr, err := spells.GetSurrenderSpell(wavinghands.GetSpell("Surrender"))
				if err != nil {
					return err, true
				}
				surrenderString, err := sr.Cast(&g.gData.Players[i], t)
				if err == nil && surrenderString != "" {
					response.Message = surrenderString
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}
				// Run Protection Spells

				cHW, err := spells.GetCureHeavyWoundsSpell(wavinghands.GetSpell("Cure Heavy Wounds"))
				if err != nil {
					return err, true
				}
				chwResult, err := cHW.Cast(&g.gData.Players[i], t)
				if err == nil && chwResult != "" {
					response.Message = chwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Run Damage Spells
				m, mErr := spells.GetMissileSpell(wavinghands.GetSpell("Missile"))
				if mErr != nil {
					return mErr, true
				}
				mResult, err := m.Cast(&g.gData.Players[i], t)
				if err == nil && mResult != "" {
					response.Message = mResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}
				CHW, err := spells.GetCauseHeavyWoundsSpell(wavinghands.GetSpell("Cause Heavy Wounds"))
				if err != nil {
					return err, true
				}
				chwResult, err = CHW.Cast(&g.gData.Players[i], t)
				if err == nil && chwResult != "" {
					response.Message = chwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				cHW.Clear(&p.Living)
			}

			// Run Summon Spells
			g.gData.Round += 1
		}

		winner, err := getWHWinner(g)
		if err == nil {
			response.Message = fmt.Sprintf("/echo %s \"has won the game of waving hands.\" 2", winner.Name)
			event.ResponseChannel <- response

			ClearGame(g.Channel.Id, event.Cache)
		} else {
			SetChannelGame(g.Channel.Id, g.gData, event.Cache)
		}
	}
	return nil, true
}

func announceGestures(
	p *wavinghands.Wizard,
	channel chan comms.Response,
	response comms.Response,
	gesture string,
	gesture2 string,
	target string) {
	protections := strings.Split(p.Protections, ",")
	gestureName := convertGesture(gesture)
	gestureName2 := convertGesture(gesture2)

	if !slices.Contains(protections, "invisible") {
		if target != "" {
			response.Message = fmt.Sprintf("%s %s and %s at %s", p.Name, gestureName, gestureName2, target)
		} else {
			response.Message = fmt.Sprintf("%s %s and %s", p.Name, gestureName, gestureName2)

		}
	}

	channel <- response
}

func convertGesture(gesture string) string {
	// Please supply 2 gestures for town-square (f, p, s, w, d, c {requires both hands}, stab, nothing)
	switch gesture {
	case "p":
		return "Proffers Palm (P)"
	case "f":
		return "Wiggled Fingers (F)"
	case "s":
		return "Snaps (S)"
	case "w":
		return "Waves (W)"
	case "d":
		return "Digit Points (D)"
	case "c":
		return "Claps (C)"
	case "1":
		return "Stabs (stabs)"
	case "0":
		return "Does Nothing (nothing)"
	}
	return ""
}

func handleEmptyBody(event BotGame) (error, bool) {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
		Quit:           nil,
	}
	g, err := NewWavingHands(event)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return nil, true
	}

	if len(g.gData.Players) >= wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo %s has joined waving hands. Would you like to start?", event.sender)
	} else if len(g.gData.Players) >= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo The waving hands game is full. Would you like to start?\n")
	} else if len(g.gData.Players) < wavinghands.GetMinTeams() {
		response.Message = fmt.Sprintf("/echo %s would like to play a game of Waving Hands.\n", event.sender)
	}
	event.ResponseChannel <- response
	return nil, false
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
		} else {
			continue
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
		if len(w.Right.Get()) <= g.gData.Round && len(w.Left.Get()) <= g.gData.Round {
			return false
		}
	}
	return true
}
