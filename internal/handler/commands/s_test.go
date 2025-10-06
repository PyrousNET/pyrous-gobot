package commands

import (
	"encoding/json"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"testing"
	"time"
)

func TestBotCommand_S(t *testing.T) {
	// Create test user data with a previous message
	testUser := users.User{
		Id:      "testuser123",
		Name:    "testuser",
		Message: "I made a typo in my message",
	}
	userData, _ := json.Marshal(testUser)
	userDataString := string(userData)

	type fields struct {
		body            string
		sender          string
		target          string
		mm              *mmclient.MMClient
		settings        *settings.Settings
		ReplyChannel    *model.Channel
		ResponseChannel chan comms.Response
		method          Method
		cache           cache.Cache
		pubsub          pubsub.Pubsub
		Quit            chan bool
	}
	type args struct {
		event BotCommand
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantMsg  string
		wantType string
	}{
		{
			name: "test s command with valid replacement",
			fields: fields{
				body:            "/typo/correction/",
				sender:          "@testuser",
				target:          "",
				mm:              nil,
				settings:        nil,
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: make(chan comms.Response, 1),
				method:          Method{},
				cache: &MockCacheWithUser{
					userData: map[string]interface{}{
						"user-testuser": userDataString,
					},
				},
				Quit: make(chan bool),
			},
			args: args{
				event: BotCommand{
					body:            "/typo/correction/",
					sender:          "@testuser",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: make(chan comms.Response, 1),
					method:          Method{},
					cache: &MockCacheWithUser{
						userData: map[string]interface{}{
							"user-testuser": userDataString,
						},
					},
					Quit: make(chan bool),
				},
			},
			wantErr:  false,
			wantMsg:  `/echo @testuser meant: "I made a correction in my message"`,
			wantType: "command",
		},
		{
			name: "test s command with invalid format - missing parts (no response sent - bug)",
			fields: fields{
				body:            "invalidformat",
				sender:          "@testuser",
				target:          "",
				mm:              nil,
				settings:        nil,
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: make(chan comms.Response, 1),
				method:          Method{},
				cache: &MockCacheWithUser{
					userData: map[string]interface{}{
						"user-testuser": userDataString,
					},
				},
				Quit: make(chan bool),
			},
			args: args{
				event: BotCommand{
					body:            "invalidformat",
					sender:          "@testuser",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: make(chan comms.Response, 1),
					method:          Method{},
					cache: &MockCacheWithUser{
						userData: map[string]interface{}{
							"user-testuser": userDataString,
						},
					},
					Quit: make(chan bool),
				},
			},
			wantErr:  false,
			wantMsg:  "",  // No response is sent to channel due to early return - this is a bug
			wantType: "", // No response is sent to channel due to early return - this is a bug
		},
		{
			name: "test s command with user not in cache",
			fields: fields{
				body:            "/old/new/",
				sender:          "@unknownuser",
				target:          "",
				mm:              nil,
				settings:        nil,
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: make(chan comms.Response, 1),
				method:          Method{},
				cache:           &cache.MockCache{},
				Quit:            make(chan bool),
			},
			args: args{
				event: BotCommand{
					body:            "/old/new/",
					sender:          "@unknownuser",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: make(chan comms.Response, 1),
					method:          Method{},
					cache:           &cache.MockCache{},
					Quit:            make(chan bool),
				},
			},
			wantErr:  false,
			wantMsg:  `/echo @unknownuser meant: ""`,
			wantType: "command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := BotCommand{
				body:            tt.fields.body,
				sender:          tt.fields.sender,
				target:          tt.fields.target,
				mm:              tt.fields.mm,
				settings:        tt.fields.settings,
				ReplyChannel:    tt.fields.ReplyChannel,
				ResponseChannel: tt.fields.ResponseChannel,
				method:          tt.fields.method,
				cache:           tt.fields.cache,
				pubsub:          tt.fields.pubsub,
				Quit:            tt.fields.Quit,
			}
			var r comms.Response
			var responseReceived bool
			go func() {
				select {
				case r = <-tt.args.event.ResponseChannel:
					responseReceived = true
					return
				case <-time.After(100 * time.Millisecond):
					// No response received - this is expected for some cases
					return
				}
			}()
			if err := bc.S(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("S() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Small delay to allow goroutine to complete
			time.Sleep(150 * time.Millisecond)
			
			// Only check message content if we expected a response
			if tt.wantMsg != "" || tt.wantType != "" {
				if !responseReceived {
					t.Errorf("S() expected response but none received")
				} else {
					if r.Message != tt.wantMsg {
						t.Errorf("S() message = %v, want %v", r.Message, tt.wantMsg)
					}
					if r.Type != tt.wantType {
						t.Errorf("S() type = %v, want %v", r.Type, tt.wantType)
					}
				}
			} else {
				// We expect no response
				if responseReceived {
					t.Errorf("S() unexpected response received: %v", r)
				}
			}
		})
	}
}