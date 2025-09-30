package games

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost/server/public/model"
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

func NewWavingHands(event *BotGame) (Game, error) {
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

func StartWavingHands(event *BotGame) (Game, error) {
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
		if len(g.gData.Players) >= wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
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

func (bg BotGame) Wh(event *BotGame) error {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
		Quit:           nil,
	}

	if event.body == "" {
		hasErr := handleEmptyBody(event)
		if hasErr {
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

func handleGameWithDirective(event *BotGame, err error) (error, bool) {
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
	case "rules":
		response.Message = "/echo Waving Hands is a turn-based wizard dueling game. See the WAVING_HANDS_RULES.md file for complete rules, or use 'wh help-spells' to see available spells."
		response.Type = "dm"
		event.ResponseChannel <- response
		return nil, true
	case "status":
		g, err := GetChannelGame(event.ReplyChannel.Id, event.Cache)
		if err != nil {
			response.Type = "dm"
			response.Message = "No active game in this channel."
			event.ResponseChannel <- response
			return nil, true
		}
		
		statusMsg := fmt.Sprintf("/echo **Waving Hands Game Status - Round %d**\n", g.Round)
		for _, player := range g.Players {
			statusMsg += fmt.Sprintf("**%s**: %d HP", player.Name, player.Living.HitPoints)
			if player.Living.Wards != "" {
				statusMsg += fmt.Sprintf(" (Protected: %s)", player.Living.Wards)
			}
			statusMsg += "\n"
		}
		
		response.Message = statusMsg
		response.Type = "command"
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
			if lGesture == "nothing" {
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
				announceGestures(&p, event.ResponseChannel, response, string(rG), string(lG), p.GetTarget())
				
				// Find the target for this player's spells
				var spellTarget *wavinghands.Living
				if p.GetTarget() != "" {
					spellTarget, err = FindTarget(g, p.GetTarget())
					if err != nil {
						return err, true
					}
				} else {
					// Default to self if no target specified
					spellTarget = &p.Living
				}
				
				sr, err := spells.GetSurrenderSpell(wavinghands.GetSpell("Surrender"))
				if err != nil {
					return err, true
				}
				surrenderString, err := sr.Cast(&g.gData.Players[i], &p.Living) // Surrender always targets self
				if err == nil && surrenderString != "" {
					response.Message = surrenderString
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}
				// Run Protection Spells
				
				// Shield
				shield, err := spells.GetShieldSpell(wavinghands.GetSpell("Shield"))
				if err != nil {
					return err, true
				}
				shieldResult, err := shield.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && shieldResult != "" {
					response.Message = shieldResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Counter Spell
				counterSpell, err := spells.GetCounterSpellSpell(wavinghands.GetSpell("Counter Spell"))
				if err != nil {
					return err, true
				}
				counterResult, err := counterSpell.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && counterResult != "" {
					response.Message = counterResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Cure Heavy Wounds
				cHW, err := spells.GetCureHeavyWoundsSpell(wavinghands.GetSpell("Cure Heavy Wounds"))
				if err != nil {
					return err, true
				}
				chwResult, err := cHW.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && chwResult != "" {
					response.Message = chwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Cure Light Wounds
				cLW, err := spells.GetCureLightWoundsSpell(wavinghands.GetSpell("Cure Light Wounds"))
				if err != nil {
					return err, true
				}
				clwResult, err := cLW.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && clwResult != "" {
					response.Message = clwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Run Mental Effects Spells

				// Anti-Spell
				antiSpell, err := spells.GetAntiSpellSpell(wavinghands.GetSpell("Anti-Spell"))
				if err != nil {
					return err, true
				}
				antiResult, err := antiSpell.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && antiResult != "" {
					response.Message = antiResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Amnesia
				amnesia, err := spells.GetAmnesiaSpell(wavinghands.GetSpell("Amnesia"))
				if err != nil {
					return err, true
				}
				amnesiaResult, err := amnesia.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && amnesiaResult != "" {
					response.Message = amnesiaResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Run Damage Spells

				// Finger of Death - should go first as it's instant kill
				fod, err := spells.GetFingerOfDeathSpell(wavinghands.GetSpell("Finger of Death"))
				if err != nil {
					return err, true
				}
				fodResult, err := fod.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && fodResult != "" {
					response.Message = fodResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Cause Heavy Wounds
				CHW, err := spells.GetCauseHeavyWoundsSpell(wavinghands.GetSpell("Cause Heavy Wounds"))
				if err != nil {
					return err, true
				}
				chwResult, err = CHW.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && chwResult != "" {
					response.Message = chwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Cause Light Wounds
				CLW, err := spells.GetCauseLightWoundsSpell(wavinghands.GetSpell("Cause Light Wounds"))
				if err != nil {
					return err, true
				}
				clwResult, err = CLW.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && clwResult != "" {
					response.Message = clwResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Missile
				missile, err := spells.GetMissileSpell(wavinghands.GetSpell("Missile"))
				if err != nil {
					return err, true
				}
				missileResult, err := missile.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && missileResult != "" {
					response.Message = missileResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Stab
				stab, err := spells.GetStabSpell(wavinghands.GetSpell("Stab"))
				if err != nil {
					return err, true
				}
				stabResult, err := stab.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && stabResult != "" {
					response.Message = stabResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Run Summon Spells
				
				// Elemental
				elemental, err := spells.GetElementalSpell(wavinghands.GetSpell("Elemental"))
				if err != nil {
					return err, true
				}
				elementalResult, err := elemental.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && elementalResult != "" {
					response.Message = elementalResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}
			}

			// Clean up expired wards at end of round
			wavinghands.CleanupAllWards(g.gData.Players)
			
			g.gData.Round += 1
		}

		winner, err := getWHWinner(g)
		if err == nil {
			response.Message = fmt.Sprintf("%s has won the game of waving hands.", winner.Name)
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

func handleEmptyBody(event *BotGame) bool {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
		Quit:           nil,
	}
	if player.Name == "" {
		response.Message = "/echo You must have a name to play Waving Hands.\n"
		event.ResponseChannel <- response
		return true
	}
	g, err := NewWavingHands(event)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return true
	}

	if len(g.gData.Players) >= wavinghands.GetMinTeams() && len(g.gData.Players) <= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo %s has joined waving hands. Would you like to start?", event.sender)
	} else if len(g.gData.Players) >= wavinghands.GetMaxTeams() {
		response.Message = fmt.Sprintf("/echo The waving hands game is full. Would you like to start?\n")
	} else if len(g.gData.Players) < wavinghands.GetMinTeams() {
		response.Message = fmt.Sprintf("/echo %s would like to play a game of Waving Hands.\n", event.sender)
	}
	event.ResponseChannel <- response
	return false
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
	
	// Find the wizard by name
	for i, w := range g.gData.Players {
		if w.Name == name {
			wizard = &g.gData.Players[i]
			break
		}
	}
	
	if wizard == nil {
		return &wavinghands.Living{}, fmt.Errorf("wizard %s not found", name)
	}

	if monster == "" {
		// Target the wizard directly
		wizard.Living.Selector = selector
		return &wizard.Living, nil
	} else {
		// Target a specific monster belonging to the wizard
		for i, m := range wizard.Monsters {
			if m.Type == monster {
				wizard.Monsters[i].Living.Selector = selector
				return &wizard.Monsters[i].Living, nil
			}
		}
		return &wavinghands.Living{}, fmt.Errorf("monster %s not found for wizard %s", monster, name)
	}
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
