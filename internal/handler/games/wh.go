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
	spellContext struct {
		caster      *wavinghands.Wizard
		target      *wavinghands.Living
		casterIndex int
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
		switch {
		case event.mm == nil && event.ReplyChannel != nil:
			channel = event.ReplyChannel
		case channelName != "":
			if event.mm != nil {
				c, err := event.mm.GetChannelByName(channelName)
				if err != nil {
					return err, true
				}
				channel = c
			} else {
				return fmt.Errorf("unable to resolve channel %s", channelName), true
			}
		case event.ReplyChannel != nil:
			channel = event.ReplyChannel
		default:
			return fmt.Errorf("no channel name included"), true
		}

		response.ReplyChannelId = channel.Id
		wHGameData, err = GetChannelGame(channel.Id, event.Cache)
		if err != nil {
			return fmt.Errorf("game state: %w", err), true
		}

		g := Game{gData: wHGameData, Channel: channel}
		p, err := GetCurrentPlayer(g, name)

		if err != nil {
			return fmt.Errorf("player lookup: %w", err), true
		} else {
			currentUser, _, err := users.GetUser(p.Name, event.Cache)
			if err != nil {
				return err, true
			}

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

			rGesture, lGesture, notices := applyPreTurnEffects(p, rGesture, lGesture)
			for _, note := range notices {
				if currentUser.Id == "" {
					continue
				}
				dm := comms.Response{
					ReplyChannelId: event.ReplyChannel.Id,
					Type:           "dm",
					UserId:         currentUser.Id,
					Message:        fmt.Sprintf("/echo %s", note),
				}
				event.ResponseChannel <- dm
			}

			rightGestures := append(p.Right.Get(), rGesture[0])
			leftGestures := append(p.Left.Get(), lGesture[0])

			p.Right.Set(string(rightGestures))
			p.Left.Set(string(leftGestures))
			p.LastRight = string(rGesture[0])
			p.LastLeft = string(lGesture[0])
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

				// Remove Enchantment
				removeEnchant, err := spells.GetRemoveEnchantmentSpell(wavinghands.GetSpell("Remove Enchantment"))
				if err != nil {
					return err, true
				}
				removeResult, err := removeEnchant.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && removeResult != "" {
					response.Message = removeResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Resist Heat
				resistHeat, err := spells.GetResistHeatSpell(wavinghands.GetSpell("Resist Heat"))
				if err != nil {
					return err, true
				}
				resistHeatResult, err := resistHeat.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && resistHeatResult != "" {
					response.Message = resistHeatResult
					event.ResponseChannel <- response
				} else if err != nil {
					return err, true
				}

				// Resist Cold
				resistCold, err := spells.GetResistColdSpell(wavinghands.GetSpell("Resist Cold"))
				if err != nil {
					return err, true
				}
				resistColdResult, err := resistCold.Cast(&g.gData.Players[i], spellTarget)
				if err == nil && resistColdResult != "" {
					response.Message = resistColdResult
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

			resolveMonsterAttacks(&g, event, response)
			wavinghands.CleanupDeadMonsters(g.gData.Players)
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

func applyPreTurnEffects(
	wizard *wavinghands.Wizard,
	rightGesture string,
	leftGesture string,
) (string, string, []string) {
	notifications := []string{}

	if wavinghands.HasWard(&wizard.Living, "anti-spell") {
		wizard.Right.Set("")
		wizard.Left.Set("")
		wavinghands.RemoveWard(&wizard.Living, "anti-spell")
		notifications = append(
			notifications,
			"Anti-Spell disrupts your preparations; all prior gestures are lost and you must start new sequences this turn.",
		)
	}

	if wavinghands.HasWard(&wizard.Living, "amnesia") {
		forcedRight := wizard.LastRight
		forcedLeft := wizard.LastLeft
		if forcedRight == "" {
			forcedRight = rightGesture
		}
		if forcedLeft == "" {
			forcedLeft = leftGesture
		}
		rightGesture = forcedRight
		leftGesture = forcedLeft
		wavinghands.RemoveWard(&wizard.Living, "amnesia")
		notifications = append(
			notifications,
			fmt.Sprintf(
				"Amnesia forces you to repeat %s and %s.",
				convertGesture(forcedRight),
				convertGesture(forcedLeft),
			),
		)
	}

	return rightGesture, leftGesture, notifications
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

func resolveMonsterAttacks(g *Game, event *BotGame, response comms.Response) {
	wavinghands.CleanupDeadMonsters(g.gData.Players)
	for i := range g.gData.Players {
		attacker := &g.gData.Players[i]
		if len(attacker.Monsters) == 0 || attacker.Target == "" {
			continue
		}

		target, err := FindTarget(*g, attacker.Target)
		if err != nil {
			continue
		}

		for mi := range attacker.Monsters {
			monster := &attacker.Monsters[mi]
			if monster.Living.HitPoints <= 0 {
				continue
			}
			stats, ok := wavinghands.GetMonsterStats(monster.Type)
			if !ok {
				continue
			}

			if wavinghands.HasShield(target) {
				response.Message = fmt.Sprintf("%s's %s attack on %s was blocked by a shield.", attacker.Name, monster.Type, target.Selector)
				event.ResponseChannel <- response
				continue
			}

			if stats.Element == "fire" && wavinghands.HasWard(target, "resist-heat") {
				response.Message = fmt.Sprintf("%s shrugs off the %s's flame.", target.Selector, monster.Type)
				event.ResponseChannel <- response
				continue
			}

			if stats.Element == "cold" && wavinghands.HasWard(target, "resist-cold") {
				response.Message = fmt.Sprintf("%s shrugs off the %s's chill.", target.Selector, monster.Type)
				event.ResponseChannel <- response
				continue
			}

			target.HitPoints -= stats.Damage
			response.Message = fmt.Sprintf("%s's %s hits %s for %d damage.", attacker.Name, monster.Type, target.Selector, stats.Damage)
			event.ResponseChannel <- response
		}
	}
	wavinghands.CleanupDeadMonsters(g.gData.Players)
}
