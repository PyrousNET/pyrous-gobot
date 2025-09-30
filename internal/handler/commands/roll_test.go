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
	"regexp"
	"testing"
	"time"
)

func TestBotCommand_Roll(t *testing.T) {
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
		wantType string
		msgRegex string
	}{
		{
			name: "test roll command",
			fields: fields{
				body:            "should I take a break?",
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
					body:            "should I take a break?",
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
			wantType: "post",
			msgRegex: `@testuser rolled a [1-5] and a [1-5] for a total of ([2-9]|10)`,
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
			if err := bc.Roll(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Roll() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Small delay to allow goroutine to capture response
			time.Sleep(10 * time.Millisecond)
			if r.Type != tt.wantType {
				t.Errorf("Roll() type = %v, want %v", r.Type, tt.wantType)
			}
			if r.UserId != "testuser123" {
				t.Errorf("Roll() userId = %v, want %v", r.UserId, "testuser123")
			}
			// Test the message format with regex since dice rolls are random
			matched, err := regexp.MatchString(tt.msgRegex, r.Message)
			if err != nil {
				t.Errorf("Roll() regex error: %v", err)
			}
			if !matched {
				t.Errorf("Roll() message = %v, doesn't match regex %v", r.Message, tt.msgRegex)
			}
		})
	}
}

func TestBotCommandHelp_Roll(t *testing.T) {
	type args struct {
		request BotCommand
	}
	tests := []struct {
		name         string
		args         args
		wantHelp     string
		wantDesc     string
	}{
		{
			name: "test roll help",
			args: args{
				request: BotCommand{},
			},
			wantHelp: "Rolls two 6 sided dice for a random response to your query.\n e.g. !roll should I take a break?",
			wantDesc: "Roll some dice!",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := BotCommandHelp{}
			response := h.Roll(tt.args.request)
			if response.Help != tt.wantHelp {
				t.Errorf("Roll() help = %v, want %v", response.Help, tt.wantHelp)
			}
			if response.Description != tt.wantDesc {
				t.Errorf("Roll() description = %v, want %v", response.Description, tt.wantDesc)
			}
		})
	}
}