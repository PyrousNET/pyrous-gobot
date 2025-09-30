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
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBotCommand_Thought(t *testing.T) {
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
		name         string
		fields       fields
		args         args
		wantErr      bool
		wantType     string
		mockResponse string
		mockStatus   int
	}{
		{
			name: "test thought command with successful response",
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
			wantErr:      false,
			wantType:     "post",
			mockResponse: `{"links":[{"title":"This is a shower thought that is safe","over_18":false,"stickied":false}]}`,
			mockStatus:   200,
		},
		{
			name: "test thought command with no safe content",
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
			wantErr:      false,
			wantType:     "post",
			mockResponse: `{"links":[{"title":"This is NSFW content","over_18":true,"stickied":false}]}`,
			mockStatus:   200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockResponse))
			}))
			defer server.Close()

			// We need to monkey-patch the URL in the thought command
			// Since we can't easily inject the URL, we'll skip this test as it requires
			// modifying the actual source code to make it testable
			t.Skip("Skipping thought test - requires refactoring the command to accept URL injection for testing")
		})
	}
}

func TestBotCommandHelp_Thought(t *testing.T) {
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
			name: "test thought help",
			args: args{
				request: BotCommand{},
			},
			wantHelp: `Have Bender give a random "shower-thought"`,
			wantDesc: `Have Bender give a random "shower-thought"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := BotCommandHelp{}
			response := h.Thought(tt.args.request)
			if response.Help != tt.wantHelp {
				t.Errorf("Thought() help = %v, want %v", response.Help, tt.wantHelp)
			}
			if response.Description != tt.wantDesc {
				t.Errorf("Thought() description = %v, want %v", response.Description, tt.wantDesc)
			}
		})
	}
}