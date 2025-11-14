package games

import (
	"encoding/json"
	"fmt"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"reflect"
	"strings"
	"testing"
	"time"
)

type MCache struct {
	data map[string]interface{}
}

func (m *MCache) Put(key string, value interface{}) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
}

func (m *MCache) PutAll(m2 map[string]interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) Get(key string) (interface{}, bool, error) {
	if m.data == nil {
		m.data = make(map[string]interface{})
	}

	if key == "user-tester" {
		u, err := json.Marshal(users.User{Id: "tester-id", Name: "tester"})
		if err != nil {
			return nil, false, err
		}
		return u, true, nil
	}

	if key == "user-player1" {
		u, err := json.Marshal(users.User{Id: "player1-id", Name: "player1"})
		if err != nil {
			return nil, false, err
		}
		return u, true, nil
	}

	val, exists := m.data[key]
	return val, exists, nil
}

func (m *MCache) GetAll(keys []string) map[string]interface{} {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) Clean(key string) {
	if m.data == nil {
		return
	}
	delete(m.data, key)
}

func (m *MCache) GetKeys(prefix string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) CleanAll() {
	m.data = make(map[string]interface{})
}

func Test_handleEmptyBody(t *testing.T) {
	type args struct {
		event    BotGame
		response Response
	}
	var channel = make(chan comms.Response)
	tests := []struct {
		name        string
		args        args
		wantMessage string
		want        bool
		wantErr     error
	}{
		{
			name: "empty input",
			args: args{
				event: BotGame{
					body:            "",
					sender:          "@tester",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: channel,
					method:          Method{},
					Cache:           &MCache{},
				},
			},
			wantMessage: "/echo @tester would like to play a game of Waving Hands.\n",
			want:        false,
		},
		{
			name: "player is missing name",
			args: args{
				event: BotGame{
					body:            "",
					sender:          "test",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: channel,
					method:          Method{},
					Cache:           &MCache{},
				},
			},
			wantMessage: "/echo You must have a name to play Waving Hands.\n",
			want:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMessage comms.Response
			go func() {
				for {
					// read from the response channel
					gotMessage = <-tt.args.event.ResponseChannel
					if gotMessage != (comms.Response{}) {
						return
					}
				}
			}()
			time.Sleep(1 * time.Second)

			hasErr := handleEmptyBody(&tt.args.event)
			if !reflect.DeepEqual(hasErr, tt.want) {
				t.Errorf("handleEmptyBody() got = %v, want %v", hasErr, tt.want)
			}
			for gotMessage == (comms.Response{}) {
				time.Sleep(1 * time.Second)
			}
			// read from the response channel
			if gotMessage.Message != tt.wantMessage {
				t.Errorf("handleEmptyBody() gotMessage = %v, want %v", gotMessage.Message, tt.wantMessage)
			}
		})
	}
}

func Test_StartWavingHands_TeamSize(t *testing.T) {
	cache := &MCache{}
	channel := &model.Channel{Id: "test-channel"}

	// Set up a game with exactly the minimum number of players (2)
	players := []wavinghands.Wizard{
		{Name: "player1", Living: wavinghands.Living{HitPoints: 15}},
		{Name: "player2", Living: wavinghands.Living{HitPoints: 15}},
	}
	gameData := WHGameData{State: "starting", Players: players, Round: 0}
	SetChannelGame(channel.Id, gameData, cache)

	event := &BotGame{
		sender:       "@player1",
		ReplyChannel: channel,
		Cache:        cache,
	}

	// This should succeed with exactly minimum players (2)
	game, err := StartWavingHands(event)
	if err != nil {
		t.Errorf("StartWavingHands() with minimum players failed: %v", err)
	}

	if game.gData.State != "playing" {
		t.Errorf("StartWavingHands() should set state to 'playing', got %v", game.gData.State)
	}
}

func Test_GestureMapping(t *testing.T) {
	tests := []struct {
		name      string
		rGesture  string
		lGesture  string
		wantRight string
		wantLeft  string
	}{
		{
			name:      "stab gestures",
			rGesture:  "stab",
			lGesture:  "stab",
			wantRight: "1",
			wantLeft:  "1",
		},
		{
			name:      "nothing gestures",
			rGesture:  "nothing",
			lGesture:  "nothing",
			wantRight: "0",
			wantLeft:  "0",
		},
		{
			name:      "mixed gestures",
			rGesture:  "stab",
			lGesture:  "nothing",
			wantRight: "1",
			wantLeft:  "0",
		},
		{
			name:      "normal gestures",
			rGesture:  "p",
			lGesture:  "w",
			wantRight: "p",
			wantLeft:  "w",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the gesture mapping logic
			rGesture := tt.rGesture
			lGesture := tt.lGesture

			// Apply the same logic as in the code
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

			if rGesture != tt.wantRight {
				t.Errorf("Right gesture mapping failed: got %v, want %v", rGesture, tt.wantRight)
			}
			if lGesture != tt.wantLeft {
				t.Errorf("Left gesture mapping failed: got %v, want %v", lGesture, tt.wantLeft)
			}
		})
	}
}

func Test_FindTarget(t *testing.T) {
	// Set up a game with multiple players
	players := []wavinghands.Wizard{
		{
			Name:   "player1",
			Living: wavinghands.Living{HitPoints: 15},
			Monsters: []wavinghands.Monster{
				{Type: "goblin", Living: wavinghands.Living{HitPoints: 3}},
			},
		},
		{Name: "player2", Living: wavinghands.Living{HitPoints: 15}},
	}
	gameData := WHGameData{State: "playing", Players: players, Round: 0}
	game := Game{gData: gameData}

	tests := []struct {
		name     string
		selector string
		wantErr  bool
		wantHP   int
	}{
		{
			name:     "target player by name",
			selector: "player2",
			wantErr:  false,
			wantHP:   15,
		},
		{
			name:     "target monster",
			selector: "player1:goblin",
			wantErr:  false,
			wantHP:   3,
		},
		{
			name:     "target non-existent player",
			selector: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target, err := FindTarget(game, tt.selector)

			if tt.wantErr && err == nil {
				t.Errorf("FindTarget() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("FindTarget() unexpected error: %v", err)
			}
			if !tt.wantErr && target.HitPoints != tt.wantHP {
				t.Errorf("FindTarget() HP = %v, want %v", target.HitPoints, tt.wantHP)
			}
		})
	}
}

func Test_SpellSequenceMatching(t *testing.T) {
	// Test the spell sequence matching logic to ensure empty sh-sequence doesn't match everything

	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "wpfd"}, // Cause Heavy Wounds sequence
		Left:  wavinghands.Hand{Sequence: "abc"},  // Some other sequence
		Name:  "testWizard",
	}

	// Create spell with empty sh-sequence (like Cause Heavy Wounds)
	spell := wavinghands.Spell{
		Name:        "Test Spell",
		Sequence:    "wpfd",
		ShSequence:  "", // Empty sh-sequence should not match any left hand
		Description: "Test",
		Usage:       "Test",
		Damage:      3,
	}

	// Test right hand match (should work)
	rightMatch := len(wizard.Right.Sequence) >= len(spell.Sequence) &&
		strings.HasSuffix(wizard.Right.Sequence, spell.Sequence)
	if !rightMatch {
		t.Errorf("Right hand should match spell sequence")
	}

	// Test left hand match with empty sh-sequence (should NOT work)
	leftMatch := spell.ShSequence != "" &&
		len(wizard.Left.Sequence) >= len(spell.ShSequence) &&
		strings.HasSuffix(wizard.Left.Sequence, spell.ShSequence)
	if leftMatch {
		t.Errorf("Left hand should NOT match when sh-sequence is empty")
	}

	// Test that the old buggy logic would incorrectly match
	oldBuggyLogic := strings.HasSuffix(wizard.Left.Sequence, spell.ShSequence)
	if !oldBuggyLogic {
		t.Errorf("Old buggy logic should match (this proves the bug existed)")
	}
}

func TestFullGameSimulation_FingerOfDeath(t *testing.T) {
	cache := &MCache{}
	storeTestUser(cache, "player1")
	storeTestUser(cache, "player2")

	channel := &model.Channel{Id: "chan-1", Name: "duel-room"}
	responseChan := make(chan comms.Response, 400)

	join := func(sender string) {
		event := BotGame{
			body:            "",
			sender:          fmt.Sprintf("@%s", sender),
			ReplyChannel:    channel,
			ResponseChannel: responseChan,
			Cache:           cache,
		}
		handleEmptyBody(&event)
		collectResponses(responseChan)
	}

	join("player1")
	join("player2")

	gameKey := wavinghands.PREFIX + channel.Id
	if _, ok := cache.data[gameKey]; !ok {
		t.Fatalf("game state missing after players joined")
	}

	startEvent := &BotGame{
		body:            "start",
		sender:          "@player1",
		ReplyChannel:    channel,
		ResponseChannel: responseChan,
		Cache:           cache,
	}
	if err, _ := handleGameWithDirective(startEvent, nil); err != nil {
		t.Fatalf("start should succeed, got error: %v", err)
	}
	collectResponses(responseChan)
	if gState, err := GetChannelGame(channel.Id, cache); err != nil {
		t.Fatalf("expected game data after start: %v", err)
	} else if len(gState.Players) != 2 {
		t.Fatalf("expected 2 players after start, got %d", len(gState.Players))
	}

	submit := func(sender, right, left, target string) {
		body := fmt.Sprintf("%s %s %s %s", channel.Name, right, left, target)
		event := &BotGame{
			body:            body,
			sender:          fmt.Sprintf("@%s", sender),
			ReplyChannel:    channel,
			ResponseChannel: responseChan,
			Cache:           cache,
		}
		if err, _ := handleGameWithDirective(event, nil); err != nil {
			var keys []string
			for k := range cache.data {
				keys = append(keys, k)
			}
			t.Fatalf("gesture submission failed for %s: %v (cache keys: %v)", sender, err, keys)
		}
	}

	type turn struct {
		gesture string
		target  string
	}
	fodSequence := []turn{
		{"p", ""},
		{"w", ""},
		{"p", ""},
		{"f", ""},
		{"s", ""},
		{"s", ""},
		{"s", ""},
		{"d", "player2"},
	}
	for _, move := range fodSequence {
		submit("player1", move.gesture, "nothing", move.target)
		submit("player2", "nothing", "nothing", "")
	}

	responses := collectResponses(responseChan)
	var winnerMsg string
	for _, resp := range responses {
		if strings.Contains(resp.Message, "has won the game of waving hands") {
			winnerMsg = resp.Message
		}
	}

	if winnerMsg == "" || !strings.Contains(winnerMsg, "player1") {
		t.Fatalf("expected winner announcement for player1, got %q (responses: %+v)", winnerMsg, responses)
	}

	if _, err := GetChannelGame(channel.Id, cache); err == nil {
		t.Fatalf("expected game state to be cleared after victory")
	}
}

func TestMonsterAttackDealsDamage(t *testing.T) {
	g := Game{
		gData: WHGameData{
			Players: []wavinghands.Wizard{
				{
					Name:   "player1",
					Target: "player2",
					Monsters: []wavinghands.Monster{
						{
							Type:   "goblin",
							Damage: 1,
							Living: wavinghands.Living{Selector: "player1:goblin#1", HitPoints: 1},
						},
					},
				},
				{
					Name:   "player2",
					Living: wavinghands.Living{Selector: "player2", HitPoints: 15},
				},
			},
		},
		Channel: &model.Channel{Id: "chan"},
	}
	respChan := make(chan comms.Response, 10)
	event := &BotGame{ResponseChannel: respChan, ReplyChannel: g.Channel}
	response := comms.Response{ReplyChannelId: g.Channel.Id}

	resolveMonsterAttacks(&g, event, response)

	if g.gData.Players[1].Living.HitPoints != 14 {
		t.Fatalf("expected monster damage to reduce HP to 14, got %d", g.gData.Players[1].Living.HitPoints)
	}
}

func TestMonsterAttackBlockedByShield(t *testing.T) {
	g := Game{
		gData: WHGameData{
			Players: []wavinghands.Wizard{
				{
					Name:   "player1",
					Target: "player2",
					Monsters: []wavinghands.Monster{
						{
							Type:   "goblin",
							Damage: 1,
							Living: wavinghands.Living{Selector: "player1:goblin#1", HitPoints: 1},
						},
					},
				},
				{
					Name:   "player2",
					Living: wavinghands.Living{Selector: "player2", HitPoints: 15, Wards: "shield"},
				},
			},
		},
		Channel: &model.Channel{Id: "chan"},
	}

	respChan := make(chan comms.Response, 10)
	event := &BotGame{ResponseChannel: respChan, ReplyChannel: g.Channel}
	response := comms.Response{ReplyChannelId: g.Channel.Id}

	resolveMonsterAttacks(&g, event, response)

	if g.gData.Players[1].Living.HitPoints != 15 {
		t.Fatalf("expected shield to block monster damage")
	}
}

func TestMagicMirrorReflectsTarget(t *testing.T) {
	g := Game{
		gData: WHGameData{
			Players: []wavinghands.Wizard{
				{Name: "caster", Living: wavinghands.Living{Selector: "caster"}},
				{Name: "target", Living: wavinghands.Living{Selector: "target"}},
			},
		},
		Channel: &model.Channel{Id: "chan"},
	}
	wavinghands.AddWard(&g.gData.Players[1].Living, "magic-mirror")
	ctx := spellContext{
		caster:      &g.gData.Players[0],
		target:      &g.gData.Players[1].Living,
		casterIndex: 0,
	}
	event := &BotGame{
		ResponseChannel: make(chan comms.Response, 4),
		ReplyChannel:    g.Channel,
	}
	response := comms.Response{ReplyChannelId: g.Channel.Id}

	reflected := spellTargetWithMirror("Missile", ctx, ctx.target, response, event)
	if reflected != &g.gData.Players[0].Living {
		t.Fatalf("expected mirror to reflect onto caster")
	}
}

func TestDispelMagicClearsWards(t *testing.T) {
	g := Game{
		gData: WHGameData{
			Players: []wavinghands.Wizard{
				{
					Name:   "caster",
					Living: wavinghands.Living{Selector: "caster"},
				},
				{
					Name:   "ally",
					Living: wavinghands.Living{Selector: "ally", Wards: "shield"},
				},
			},
		},
		Channel: &model.Channel{Id: "chan"},
	}
	g.gData.Players[0].Right.Set("cdpw")
	contexts := []spellContext{
		{caster: &g.gData.Players[0], target: &g.gData.Players[0].Living, casterIndex: 0},
		{caster: &g.gData.Players[1], target: &g.gData.Players[1].Living, casterIndex: 1},
	}
	event := &BotGame{
		ResponseChannel: make(chan comms.Response, 4),
		ReplyChannel:    g.Channel,
	}
	response := comms.Response{ReplyChannelId: g.Channel.Id}

	triggered := resolveDispelMagic(&g, contexts, event, response)
	if !triggered {
		t.Fatalf("expected dispel magic to trigger")
	}
	for _, player := range g.gData.Players {
		if player.Name == "caster" {
			if player.Living.Wards != "shield" {
				t.Fatalf("expected caster to retain shield, got %s", player.Living.Wards)
			}
		} else if player.Living.Wards != "" {
			t.Fatalf("expected wards cleared, got %s", player.Living.Wards)
		}
	}
}

func TestFormatMonsters(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Name:   "player",
		Target: "enemy",
		Monsters: []wavinghands.Monster{
			{Type: "goblin", Damage: 1, Living: wavinghands.Living{HitPoints: 1}},
		},
	}
	text := formatMonsters(wizard)
	if !strings.Contains(text, "Goblin") {
		t.Fatalf("expected goblin in summary: %s", text)
	}
	if !strings.Contains(text, "Target: enemy") {
		t.Fatalf("expected target summary: %s", text)
	}
}
func TestMonsterSummoningFlow(t *testing.T) {
	cache := &MCache{}
	storeTestUser(cache, "player1")
	storeTestUser(cache, "player2")
	channel := &model.Channel{Id: "chan-2", Name: "arena"}
	respChan := make(chan comms.Response, 200)
	join := func(sender string) {
		event := BotGame{
			body:            "",
			sender:          fmt.Sprintf("@%s", sender),
			ReplyChannel:    channel,
			ResponseChannel: respChan,
			Cache:           cache,
		}
		handleEmptyBody(&event)
	}
	join("player1")
	join("player2")
	startEvent := &BotGame{
		body:            "start",
		sender:          "@player1",
		ReplyChannel:    channel,
		ResponseChannel: respChan,
		Cache:           cache,
	}
	if err, _ := handleGameWithDirective(startEvent, nil); err != nil {
		t.Fatalf("start failed: %v", err)
	}
	collectResponses(respChan)
	submit := func(sender, right, left, target string) {
		event := &BotGame{
			body:            fmt.Sprintf("%s %s %s %s", channel.Name, right, left, target),
			sender:          fmt.Sprintf("@%s", sender),
			ReplyChannel:    channel,
			ResponseChannel: respChan,
			Cache:           cache,
		}
		if err, _ := handleGameWithDirective(event, nil); err != nil {
			t.Fatalf("gesture submission failed: %v", err)
		}
	}
	submit("player1", "p", "s", "player2")
	submit("player2", "nothing", "nothing", "")
	submit("player1", "s", "f", "player2")
	submit("player2", "nothing", "nothing", "")
	submit("player1", "f", "w", "player2")
	submit("player2", "nothing", "nothing", "")
	submit("player1", "w", "p", "player2")
	submit("player2", "nothing", "nothing", "")
	submit("player1", "sfw", "nothing", "player2")
	submit("player2", "nothing", "nothing", "")
	responses := collectResponses(respChan)
	var found bool
	for _, r := range responses {
		if strings.Contains(r.Message, "summons a goblin") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected goblin summon, responses: %+v", responses)
	}
	g, err := GetChannelGame(channel.Id, cache)
	if err != nil {
		t.Fatalf("expected game state: %v", err)
	}
	goblinFound := false
	for _, m := range g.Players[0].Monsters {
		if m.Type == "goblin" {
			goblinFound = true
		}
	}
	if !goblinFound {
		t.Fatalf("expected player1 to have a goblin")
	}
}

func storeTestUser(cache *MCache, username string) {
	user := users.User{
		Id:   fmt.Sprintf("%s-id", username),
		Name: username,
	}
	data, _ := json.Marshal(user)
	cache.Put(users.KeyPrefix+username, data)
}

func collectResponses(ch chan comms.Response) []comms.Response {
	var responses []comms.Response
	for {
		select {
		case resp := <-ch:
			responses = append(responses, resp)
		default:
			return responses
		}
	}
}
func TestApplyPreTurnEffectsAntiSpell(t *testing.T) {
	wizard := &wavinghands.Wizard{
		Right: wavinghands.Hand{Sequence: "wpf"},
		Left:  wavinghands.Hand{Sequence: "sd"},
		Living: wavinghands.Living{
			Wards: "anti-spell",
		},
	}

	right, left, notes := applyPreTurnEffects(wizard, "p", "w")

	if wizard.Right.Sequence != "" || wizard.Left.Sequence != "" {
		t.Fatalf("expected gesture history to reset, got %s / %s", wizard.Right.Sequence, wizard.Left.Sequence)
	}
	if wavinghands.HasWard(&wizard.Living, "anti-spell") {
		t.Fatalf("anti-spell ward should be cleared after applying effect")
	}
	if right != "p" || left != "w" {
		t.Fatalf("anti-spell should not alter requested gestures, got %s / %s", right, left)
	}
	if len(notes) == 0 || !strings.Contains(notes[0], "Anti-Spell") {
		t.Fatalf("expected anti-spell notification, got %#v", notes)
	}
}

func TestApplyPreTurnEffectsAmnesia(t *testing.T) {
	wizard := &wavinghands.Wizard{
		LastRight: "p",
		LastLeft:  "w",
		Living: wavinghands.Living{
			Wards: "amnesia",
		},
	}

	right, left, notes := applyPreTurnEffects(wizard, "s", "d")

	if right != "p" || left != "w" {
		t.Fatalf("expected gestures to be forced to previous values, got %s / %s", right, left)
	}
	if wavinghands.HasWard(&wizard.Living, "amnesia") {
		t.Fatalf("amnesia ward should be cleared after forcing gestures")
	}
	if len(notes) == 0 || !strings.Contains(notes[0], "Amnesia") {
		t.Fatalf("expected amnesia notification, got %#v", notes)
	}
}
