package games

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"reflect"
	"testing"
)

func Test_handleEmptyBody(t *testing.T) {
	type args struct {
		event    BotGame
		response Response
	}
	rs := make(chan comms.Response)
	tests := []struct {
		name  string
		args  args
		want  error
		want1 bool
		want2 comms.Response
	}{
		{
			name: "empty input",
			args: args{
				event: BotGame{
					body:            "",
					sender:          "",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: rs,
					method:          Method{},
					Cache:           &cache.MockCache{},
				},
			},
			want:  nil,
			want1: true,
			want2: comms.Response{
				ReplyChannelId: "test",
				Message:        "player is missing a name",
				Type:           "dm",
				UserId:         "",
				Quit:           nil,
			},
		},
		{
			name: "test body",
			args: args{
				event: BotGame{
					body:            "",
					sender:          "test",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: rs,
					method:          Method{},
					Cache:           &cache.MockCache{},
				},
			},
			want:  nil,
			want1: false,
			want2: comms.Response{
				ReplyChannelId: "test",
				Message:        "/echo test would like to play a game of Waving Hands.\n",
				Type:           "command",
				UserId:         "",
				Quit:           nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				got, got1 := handleEmptyBody(tt.args.event)
				if got != tt.want {
					t.Errorf("handleEmptyBody() got = %v, want %v", got1, tt.want1)
				}

				if got1 != tt.want1 {
					t.Errorf("handleEmptyBody() got1 = %v, want1 %v", got1, tt.want1)
				}
			}()
			got2 := <-rs
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("handleEmptyBody() got2 = %v, want2 %v", got2, tt.want2)
			}
		})
	}
}
