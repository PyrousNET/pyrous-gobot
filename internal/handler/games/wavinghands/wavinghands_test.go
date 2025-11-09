package wavinghands

import "testing"

func TestCleanupWards(t *testing.T) {
	living := &Living{
		Selector:  "test",
		HitPoints: 15,
		Wards:     "shield,counter-spell",
	}

	CleanupWards(living)

	if living.Wards != "" {
		t.Errorf("Expected wards to be empty after cleanup, got '%s'", living.Wards)
	}
}

func TestCleanupWards_Persistent(t *testing.T) {
	living := &Living{
		Selector:  "test",
		HitPoints: 15,
		Wards:     "shield,amnesia",
	}

	CleanupWards(living)

	if living.Wards != "amnesia" {
		t.Errorf("Expected amnesia ward to persist, got '%s'", living.Wards)
	}
}

func TestCleanupAllWards(t *testing.T) {
	players := []Wizard{
		{
			Name: "Player1",
			Living: Living{
				Selector:  "player1",
				HitPoints: 15,
				Wards:     "shield",
			},
		},
		{
			Name: "Player2",
			Living: Living{
				Selector:  "player2",
				HitPoints: 12,
				Wards:     "counter-spell,cure-heavy-wounds",
			},
		},
	}

	CleanupAllWards(players)

	for i, player := range players {
		if player.Living.Wards != "" {
			t.Errorf("Expected player %d wards to be empty after cleanup, got '%s'", i, player.Living.Wards)
		}
	}
}

func TestGetMaxTeams(t *testing.T) {
	max := GetMaxTeams()
	if max != 6 {
		t.Errorf("Expected max teams to be 6, got %d", max)
	}
}

func TestGetSpellSequences(t *testing.T) {
	tests := []struct {
		name     string
		sequence string
	}{
		{"Anti-Spell", "spf"},
		{"Counter Spell", "wpp|wws"},
		{"Finger of Death", "pwpfsssd"},
		{"Surrender", "p"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spell, err := GetSpell(tt.name)
			if err != nil {
				t.Fatalf("GetSpell(%s) error: %v", tt.name, err)
			}
			if spell.Sequence != tt.sequence {
				t.Fatalf("spell %s sequence mismatch: got %q want %q", tt.name, spell.Sequence, tt.sequence)
			}
		})
	}
}

func TestGetMinTeams(t *testing.T) {
	min := GetMinTeams()
	if min != 2 {
		t.Errorf("Expected min teams to be 2, got %d", min)
	}
}

func TestFormatWards(t *testing.T) {
	living := Living{Wards: "shield,protection-from-evil:3,resist-heat"}
	got := FormatWards(living)
	expected := "Shield, Protection from Evil (3), Resist Heat"
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
