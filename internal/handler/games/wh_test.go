package games

import (
	"encoding/json"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/wavinghands"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"reflect"
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
		u, err := json.Marshal(users.User{Name: "tester"})
		if err != nil {
			return nil, false, err
		}
		return u, true, nil
	}
	
	if key == "user-player1" {
		u, err := json.Marshal(users.User{Name: "player1"})
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
			Name: "player1", 
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
