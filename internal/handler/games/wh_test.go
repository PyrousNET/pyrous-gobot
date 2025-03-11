package games

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"reflect"
	"testing"
	"time"
)

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
					Cache:           &cache.MockCache{},
				},
			},
			wantMessage: "/echo @tester would like to play a game of Waving Hands.\n",
			want:        false,
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
