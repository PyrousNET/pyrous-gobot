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

func TestBotCommand_Emote(t *testing.T) {
	// Create test user data
	testUser := users.User{
		Id:      "testuser123",
		Name:    "testuser",
		Message: "test message",
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
			name: "test emote command with text",
			fields: fields{
				body:            "waves hello",
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
					body:            "waves hello",
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
			wantMsg:  `/me waves hello`,
			wantType: "command",
		},
		{
			name: "test emote command with empty body",
			fields: fields{
				body:            "",
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
					body:            "",
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
			wantMsg:  `/me `,
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
			go func() {
				select {
				case r = <-tt.args.event.ResponseChannel:
					return
				case <-time.After(2 * time.Second):
					t.Error("Test timed out waiting for response")
					return
				}
			}()
			if err := bc.Emote(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Emote() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Small delay to allow goroutine to capture response
			time.Sleep(10 * time.Millisecond)
			if r.Message != tt.wantMsg {
				t.Errorf("Emote() message = %v, want %v", r.Message, tt.wantMsg)
			}
			if r.Type != tt.wantType {
				t.Errorf("Emote() type = %v, want %v", r.Type, tt.wantType)
			}
			if r.UserId != "testuser123" {
				t.Errorf("Emote() userId = %v, want %v", r.UserId, "testuser123")
			}
		})
	}
}