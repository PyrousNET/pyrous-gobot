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

func TestGetMinTeams(t *testing.T) {
	min := GetMinTeams()
	if min != 2 {
		t.Errorf("Expected min teams to be 2, got %d", min)
	}
}