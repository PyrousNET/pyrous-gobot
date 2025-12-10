package games

import "testing"

func TestScoreSelection(t *testing.T) {
	tests := []struct {
		name    string
		dice    []int
		want    int
		wantErr bool
	}{
		{"single one", []int{1}, 100, false},
		{"single five", []int{5}, 50, false},
		{"triple twos", []int{2, 2, 2}, 200, false},
		{"four twos", []int{2, 2, 2, 2}, 400, false},
		{"five ones", []int{1, 1, 1, 1, 1}, 3000, false},
		{"straight", []int{1, 2, 3, 4, 5, 6}, 1500, false},
		{"three pairs", []int{1, 1, 2, 2, 3, 3}, 1500, false},
		{"two triplets", []int{2, 2, 2, 3, 3, 3}, 2500, false},
		{"four kind and pair", []int{4, 4, 4, 4, 5, 5}, 1500, false},
		{"invalid selection", []int{2, 3}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := scoreSelection(tt.dice)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("scoreSelection(%v) = %d, want %d", tt.dice, got, tt.want)
			}
		})
	}
}

func TestHasScoringDice(t *testing.T) {
	if hasScoringDice([]int{2, 3, 4, 6}) {
		t.Fatalf("expected no scoring dice")
	}
	if !hasScoringDice([]int{1, 2, 3}) {
		t.Fatalf("expected scoring dice with a single 1")
	}
	if !hasScoringDice([]int{2, 2, 2}) {
		t.Fatalf("expected scoring dice with a triple")
	}
	if !hasScoringDice([]int{1, 2, 3, 4, 5, 6}) {
		t.Fatalf("expected scoring dice with a straight")
	}
}

func TestIsSubset(t *testing.T) {
	roll := []int{1, 2, 3, 4, 5, 6}
	if !isSubset([]int{1, 5}, roll) {
		t.Fatalf("expected subset to be true")
	}
	if isSubset([]int{1, 1}, roll) {
		t.Fatalf("expected subset to be false when exceeding counts")
	}
}

func TestBestScoringSubsetCountMode(t *testing.T) {
	roll := []int{1, 5, 5, 2, 3, 4}

	subset, warn, err := bestScoringSubset(roll, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warn == "" {
		t.Fatalf("expected warning when requesting more scoring dice than available")
	}
	if len(subset) != 3 {
		t.Fatalf("expected 3 scoring dice kept, got %v", subset)
	}
	score, err := scoreSelection(subset)
	if err != nil || score == 0 {
		t.Fatalf("subset should be scoring, score=%d err=%v", score, err)
	}

	// Keeping 1 should pick the best single scoring die (the 1 worth 100 over a 5 worth 50).
	subset, warn, err = bestScoringSubset(roll, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if warn != "" {
		t.Fatalf("did not expect warning for exact available dice")
	}
	if len(subset) != 1 || subset[0] != 1 {
		t.Fatalf("expected to keep the single '1' die, got %v", subset)
	}

	// Prefer higher value scoring when scores tie: choose 1+5 (150) over two 5s (100).
	subset, _, err = bestScoringSubset(roll, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	score, err = scoreSelection(subset)
	if err != nil {
		t.Fatalf("score error: %v", err)
	}
	if score != 150 {
		t.Fatalf("expected highest scoring dice kept (score 150), got %d with %v", score, subset)
	}
}
