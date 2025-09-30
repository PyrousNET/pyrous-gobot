package commands

import "testing"

func TestConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"Err constant", Err, -1},
		{"Say constant", Say, 0},
		{"Emote constant", Emote, 1},
		{"Reply constant", Reply, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}