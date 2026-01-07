package commands

import (
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"sync"
	"testing"
)

func TestBotCommand_React(t *testing.T) {
	sttngs := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
		Reactions: map[string]settings.Reaction{
			"test": {
				Description: "test",
				Url:         "test",
			},
		},
	})
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
			name: "test simple react formatted as image",
			fields: fields{
				body:            "test",
				sender:          "@test",
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
					body:            "test",
					sender:          "@test",
					target:          "",
					mm:              nil,
					settings:        sttngs,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: make(chan comms.Response, 1),
					method:          Method{},
					cache:           &cache.MockCache{},
					Quit:            make(chan bool),
				},
			},
			wantErr: false,
			wantMsg: "/echo \"![test](test)\" 1",
		},
		{
			name: "test react is case-insensitive with normalized spacing",
			fields: fields{
				body:            "TeSt   ",
				sender:          "@test",
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
					body:            "TeSt   ",
					sender:          "@test",
					target:          "",
					mm:              nil,
					settings:        sttngs,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: make(chan comms.Response, 1),
					method:          Method{},
					cache:           &cache.MockCache{},
					Quit:            make(chan bool),
				},
			},
			wantErr: false,
			wantMsg: "/echo \"![test](test)\" 1",
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
			if err := bc.React(tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("React() error = %v, wantErr %v", err, tt.wantErr)
			}
			r = <-tt.args.event.ResponseChannel
			if r.Message != tt.wantMsg {
				t.Errorf("React() = %v, want %v", r.Message, tt.wantMsg)
			}
		})
	}
}
