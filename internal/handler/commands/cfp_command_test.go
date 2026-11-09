package commands

import (
	"strings"
	"sync"
	"testing"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestCFPCommandIsRegistered(t *testing.T) {
	commands := NewCommands(settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{CommandTrigger: "!"}), nil, &cache.MockCache{}, nil)
	command, err := commands.NewBotCommand("!cfp", "test")
	if err != nil {
		t.Fatalf("NewBotCommand(!cfp) error = %v", err)
	}
	if command.target != "cfp" {
		t.Fatalf("command target = %q, want cfp", command.target)
	}
	doc := (BotCommandHelp{}).Cfp(BotCommand{})
	if !strings.Contains(doc.Help, "!cfp") {
		t.Fatalf("CFP help = %q, want !cfp usage", doc.Help)
	}
}
