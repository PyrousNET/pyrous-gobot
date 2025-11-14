package commands

import (
	"sync"
	"testing"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestNewBotCommandDiceMapped(t *testing.T) {
	tests := []struct {
		name           string
		trigger        string
		command        string
		expectedTarget string
	}{
		{
			name:           "simple trigger",
			trigger:        "!",
			command:        "!2d20 roll initiative",
			expectedTarget: "2d20",
		},
		{
			name:           "trigger with trailing space",
			trigger:        "! ",
			command:        "!dice 2d6 escape",
			expectedTarget: "dice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
				CommandTrigger: tt.trigger,
			})

			cmds := NewCommands(st, nil, &cache.MockCache{}, nil)
			bc, err := cmds.NewBotCommand(tt.command, "@adventurer")
			if err != nil {
				t.Fatalf("NewBotCommand returned error: %v", err)
			}

			if bc.method.typeOf.Name != "Dice" {
				t.Fatalf("expected Dice method, got %s", bc.method.typeOf.Name)
			}

			if bc.target != tt.expectedTarget {
				t.Fatalf("expected target %s, got %s", tt.expectedTarget, bc.target)
			}
		})
	}
}
