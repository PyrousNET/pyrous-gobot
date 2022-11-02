package games

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"reflect"
	"testing"
)

func Test_handleEmptyBody(t *testing.T) {
	type args struct {
		event    BotGame
		response Response
	}
	tests := []struct {
		name  string
		args  args
		want  Response
		want1 error
		want2 bool
	}{
		{
			name: "empty input",
			args: args{
				event: BotGame{
					body:         "",
					sender:       "",
					target:       "",
					mm:           nil,
					settings:     nil,
					ReplyChannel: &model.Channel{Id: "test"},
					method:       Method{},
					cache:        &cache.MockCache{},
				},
			},
			want:  Response{},
			want1: nil,
			want2: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := handleEmptyBody(tt.args.event, tt.args.response)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleEmptyBody() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("handleEmptyBody() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("handleEmptyBody() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}
