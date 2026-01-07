package commands

import (
	"strings"
	"sync"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestBotCommand_Help_LookupIsCaseInsensitive(t *testing.T) {
	sttngs := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
		Reactions: map[string]settings.Reaction{
			"test": {
				Description: "test",
				Url:         "test",
			},
		},
	})

	respCh := make(chan comms.Response, 1)
	event := BotCommand{
		body:            "ReAcT",
		sender:          "@test",
		settings:        sttngs,
		ReplyChannel:    &model.Channel{Id: "test"},
		ResponseChannel: respCh,
		cache:           &cache.MockCache{},
	}

	if err := event.Help(event); err != nil {
		t.Fatalf("Help error: %v", err)
	}

	resp := <-respCh
	if !strings.Contains(resp.Message, "test - test") {
		t.Fatalf("unexpected help response: %q", resp.Message)
	}
}
