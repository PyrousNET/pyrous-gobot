package commands

import (
	"sync"
	"testing"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestNewBotCommandDiceMapped(t *testing.T) {
	st := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
		CommandTrigger: "!",
	})

	cmds := NewCommands(st, nil, &cache.MockCache{}, nil)
	bc, err := cmds.NewBotCommand("!2d20 roll initiative", "@adventurer")
	if err != nil {
		t.Fatalf("NewBotCommand returned error: %v", err)
	}

	if bc.method.typeOf.Name != "Dice" {
		t.Fatalf("expected Dice method, got %s", bc.method.typeOf.Name)
	}

	if bc.target != "2d20" {
		t.Fatalf("expected target 2d20, got %s", bc.target)
	}

	if bc.body != "roll initiative" {
		t.Fatalf("expected body 'roll initiative', got %s", bc.body)
	}
}
