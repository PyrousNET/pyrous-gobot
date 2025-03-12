package games

import (
	"encoding/json"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	"reflect"
	"testing"
	"time"
)

type MCache cache.MockCache

func (m *MCache) Put(key string, value interface{}) {
	return
}

func (m *MCache) PutAll(m2 map[string]interface{}) {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) Get(key string) (interface{}, bool, error) {
	if key == "user-tester" {
		u, err := json.Marshal(users.User{Name: "tester"})
		if err != nil {
			return nil, false, err
		}
		return u, true, nil
	}
	return nil, false, nil
}

func (m *MCache) GetAll(keys []string) map[string]interface{} {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) Clean(key string) {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) GetKeys(prefix string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MCache) CleanAll() {
	//TODO implement me
	panic("implement me")
}

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
					Cache:           &MCache{},
				},
			},
			wantMessage: "/echo @tester would like to play a game of Waving Hands.\n",
			want:        false,
		},
		{
			name: "player is missing name",
			args: args{
				event: BotGame{
					body:            "",
					sender:          "test",
					target:          "",
					mm:              nil,
					settings:        nil,
					ReplyChannel:    &model.Channel{Id: "test"},
					ResponseChannel: channel,
					method:          Method{},
					Cache:           &MCache{},
				},
			},
			wantMessage: "/echo You must have a name to play Waving Hands.\n",
			want:        true,
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
