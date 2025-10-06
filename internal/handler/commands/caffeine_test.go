package commands

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"testing"
	"time"
)

func TestBotCommand_Caffeine(t *testing.T) {
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
		name    string
		fields  fields
		args    args
		wantErr bool
		wantMsg string
	}{
		{
			name: "test caffeine without number - default message",
			fields: fields{
				body:            "",
				sender:          "@testuser",
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
					body:            "",
					sender:          "@testuser",
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
			wantErr: false,
			wantMsg: "/me walks over to @testuser and gives them a shot of caffeine straight into the blood stream.",
		},
		{
			name: "test caffeine with number 5",
			fields: fields{
				body:            "5",
				sender:          "@testuser",
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
					body:            "5",
					sender:          "@testuser",
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
			wantErr: false,
			wantMsg: "/me walks over to @testuser and gives them 5 shots of caffeine straight into the blood stream.",
		},
		{
			name: "test caffeine with invalid input - no message set (potential bug)",
			fields: fields{
				body:            "abc",
				sender:          "@testuser",
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
					body:            "abc",
					sender:          "@testuser",
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
			wantErr: false,
			wantMsg: "", // Currently, the command doesn't set a message for invalid input
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
			if err := bc.Caffeine(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("Caffeine() error = %v, wantErr %v", err, tt.wantErr)
			}
			// Small delay to allow goroutine to capture response
			time.Sleep(10 * time.Millisecond)
			if r.Message != tt.wantMsg {
				t.Errorf("Caffeine() = %v, want %v", r.Message, tt.wantMsg)
			}
		})
	}
}