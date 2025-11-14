package commands

import (
	"math/rand"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
)

func TestBotCommand_Roll(t *testing.T) {
	originalRand := rollRand
	defer func() {
		rollRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}()

	tests := []struct {
		name        string
		sender      string
		seed        int64
		wantMessage string
	}{
		{
			name:        "basic roll",
			sender:      "@roller",
			seed:        1,
			wantMessage: "@roller rolled a 2 and a 3 for a total of 5",
		},
		{
			name:        "different user",
			sender:      "@wizard",
			seed:        7,
			wantMessage: "@wizard rolled a 2 and a 1 for a total of 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollRand = rand.New(rand.NewSource(tt.seed))

			responseCh := make(chan comms.Response, 1)
			event := BotCommand{
				body:            "",
				sender:          tt.sender,
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: responseCh,
				cache:           &cache.MockCache{},
			}

			bc := BotCommand{cache: &cache.MockCache{}}

			if err := bc.Roll(event); err != nil {
				t.Fatalf("Roll() error = %v", err)
			}

			resp := <-responseCh
			if resp.Message != tt.wantMessage {
				t.Fatalf("Roll() = %s, want %s", resp.Message, tt.wantMessage)
			}
		})
	}

	rollRand = originalRand
}
