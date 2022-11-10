package games

import (
	"fmt"
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
		},
		{
			name: "",
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
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				got2 := <-rs
				fmt.Println(got2)
			}()
			got, got1 := handleEmptyBody(tt.args.event)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEmptyBody() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("handleEmptyBody() got2 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
