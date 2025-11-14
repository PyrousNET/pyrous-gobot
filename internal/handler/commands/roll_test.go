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
		seed        int64
		body        string
		wantMessage string
	}{
		{
			name:        "default roll",
			seed:        1,
			body:        "",
			wantMessage: "@roller rolled a 6 and a 4 for a total of 10",
		},
		{
			name:        "default roll with reason",
			seed:        1,
			body:        "should I nap?",
			wantMessage: "@roller rolled a 6 and a 4 for a total of 10 - should I nap?",
		},
		{
			name:        "single die NdM",
			seed:        2,
			body:        "1d20 attack",
			wantMessage: "@roller rolled 1d20 and got 7 - attack",
		},
		{
			name:        "multiple dice NdM",
			seed:        7,
			body:        "3d4 escape",
			wantMessage: "@roller rolled 3d4 (3 + 3 + 2) for a total of 8 - escape",
		},
		{
			name:        "too many dice",
			seed:        1,
			body:        "51d6",
			wantMessage: "that's too many dice! Please roll 50 or fewer.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rollRand = rand.New(rand.NewSource(tt.seed))

			responseCh := make(chan comms.Response, 1)
			event := BotCommand{
				body:            tt.body,
				sender:          "@roller",
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
