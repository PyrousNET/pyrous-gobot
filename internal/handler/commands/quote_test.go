package commands

import (
	"sync"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestBotCommand_Quote(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		quotes   []string
		wantMsg  string
		contains bool
	}{
		{
			name:    "no quotes configured",
			body:    "",
			quotes:  nil,
			wantMsg: "No quotes are configured.",
		},
		{
			name:    "invalid index",
			body:    "nope",
			quotes:  []string{"first"},
			wantMsg: "Invalid quote index. Use !quote or !quote <number>.",
		},
		{
			name:    "index out of range",
			body:    "5",
			quotes:  []string{"first"},
			wantMsg: "Quote 5 not found.",
		},
		{
			name:     "specific quote",
			body:     "0",
			quotes:   []string{"hello {0}"},
			wantMsg:  "hello @tester",
			contains: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sttngs := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
				Quotes: tt.quotes,
			})
			respCh := make(chan comms.Response, 1)

			event := BotCommand{
				body:            tt.body,
				sender:          "@tester",
				settings:        sttngs,
				ReplyChannel:    &model.Channel{Id: "test"},
				ResponseChannel: respCh,
				cache:           &cache.MockCache{},
			}

			if err := event.Quote(event); err != nil {
				t.Fatalf("Quote error: %v", err)
			}

			resp := <-respCh
			if tt.contains {
				if resp.Message != tt.wantMsg {
					t.Fatalf("got %q, want %q", resp.Message, tt.wantMsg)
				}
			} else if resp.Message != tt.wantMsg {
				t.Fatalf("got %q, want %q", resp.Message, tt.wantMsg)
			}
		})
	}
}
