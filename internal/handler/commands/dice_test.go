package commands

import (
	"math/rand"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
)

func TestBotCommand_Dice(t *testing.T) {
	originalRand := diceRand
	defer func() {
		diceRand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}()

	tests := []struct {
		name        string
		target      string
		body        string
		seed        int64
		wantMessage string
		wantErr     bool
	}{
		{
			name:        "single die roll",
			target:      "1d20",
			body:        "",
			seed:        1,
			wantMessage: "@adventurer rolled 1d20 and got 2",
		},
		{
			name:        "multiple dice with reason",
			target:      "3d4",
			body:        "initiative",
			seed:        7,
			wantMessage: "@wizard rolled 3d4 (3 + 3 + 2) for a total of 8 - initiative",
		},
		{
			name:    "invalid dice format",
			target:  "xd20",
			body:    "",
			seed:    1,
			wantErr: true,
		},
		{
			name:    "too many dice",
			target:  "100d6",
			body:    "",
			seed:    1,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diceRand = rand.New(rand.NewSource(tt.seed))

			responseCh := make(chan comms.Response, 1)
			event := BotCommand{
				target:          tt.target,
				body:            tt.body,
				sender:          "@adventurer",
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: responseCh,
				cache:           &cache.MockCache{},
			}

			if tt.name == "multiple dice with reason" {
				event.sender = "@wizard"
			}

			bc := BotCommand{
				cache: &cache.MockCache{},
			}

			err := bc.Dice(event)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Dice() expected error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Dice() unexpected error = %v", err)
			}

			resp := <-responseCh
			if resp.Message != tt.wantMessage {
				t.Fatalf("Dice() = %s, want %s", resp.Message, tt.wantMessage)
			}
		})
	}

	diceRand = originalRand
}
