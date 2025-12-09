package games

import (
	"fmt"
	"strconv"
	"strings"
)

func parseDiceSelection(args []string) ([]int, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("specify dice to keep, e.g. `$farkle keep 1 5 5`")
	}
	var dice []int
	for _, a := range args {
		val, err := strconv.Atoi(a)
		if err != nil {
			return nil, fmt.Errorf("invalid die value %q", a)
		}
		if val < 1 || val > 6 {
			return nil, fmt.Errorf("die values must be between 1 and 6")
		}
		dice = append(dice, val)
	}
	return dice, nil
}

func isSubset(keep, roll []int) bool {
	rollCount := make(map[int]int)
	for _, d := range roll {
		rollCount[d]++
	}
	for _, d := range keep {
		rollCount[d]--
		if rollCount[d] < 0 {
			return false
		}
	}
	return true
}

func hasScoringDice(dice []int) bool {
	counts := make([]int, 7)
	for _, d := range dice {
		counts[d]++
	}
	// Singles
	if counts[1] > 0 || counts[5] > 0 {
		return true
	}
	// Triples or better
	for i := 1; i <= 6; i++ {
		if counts[i] >= 3 {
			return true
		}
	}
	// Straights / three pairs / two triplets / four-kind + pair
	if len(dice) == 6 {
		if isStraight(counts) || isThreePairs(counts) || isTwoTriplets(counts) || isFourKindAndPair(counts) {
			return true
		}
	}
	return false
}

func scoreSelection(dice []int) (int, error) {
	if len(dice) == 0 {
		return 0, fmt.Errorf("no dice selected")
	}
	counts := make([]int, 7)
	for _, d := range dice {
		if d < 1 || d > 6 {
			return 0, fmt.Errorf("invalid die value %d", d)
		}
		counts[d]++
	}

	// Special 6-dice combos
	if len(dice) == 6 {
		if isStraight(counts) {
			return 1500, nil
		}
		if isThreePairs(counts) {
			return 1500, nil
		}
		if isTwoTriplets(counts) {
			return 2500, nil
		}
		if isFourKindAndPair(counts) {
			return 1500, nil
		}
	}

	score := 0

	for face := 1; face <= 6; face++ {
		if counts[face] >= 3 {
			base := 1000
			if face != 1 {
				base = face * 100
			}
			multiplier := counts[face] - 2 // 3=>1x, 4=>2x, 5=>3x, 6=>4x
			score += base * multiplier
			counts[face] = 0
		}
	}

	// Singles for 1s and 5s
	if counts[1] > 0 {
		score += counts[1] * 100
		counts[1] = 0
	}
	if counts[5] > 0 {
		score += counts[5] * 50
		counts[5] = 0
	}

	// Any remaining dice are non-scoring
	for face := 2; face <= 6; face++ {
		if counts[face] > 0 {
			return 0, fmt.Errorf("selection contains unscorable dice")
		}
	}

	if score == 0 {
		return 0, fmt.Errorf("no scoring dice selected")
	}

	return score, nil
}

func isStraight(counts []int) bool {
	for i := 1; i <= 6; i++ {
		if counts[i] != 1 {
			return false
		}
	}
	return true
}

func isThreePairs(counts []int) bool {
	pairs := 0
	for i := 1; i <= 6; i++ {
		if counts[i] == 2 {
			pairs++
		} else if counts[i] != 0 {
			return false
		}
	}
	return pairs == 3
}

func isTwoTriplets(counts []int) bool {
	triplets := 0
	for i := 1; i <= 6; i++ {
		if counts[i] == 3 {
			triplets++
		} else if counts[i] != 0 {
			return false
		}
	}
	return triplets == 2
}

func isFourKindAndPair(counts []int) bool {
	foundFour := false
	foundPair := false
	for i := 1; i <= 6; i++ {
		if counts[i] == 4 {
			foundFour = true
		} else if counts[i] == 2 {
			foundPair = true
		} else if counts[i] != 0 {
			return false
		}
	}
	return foundFour && foundPair
}

func formatDice(dice []int) string {
	parts := make([]string, len(dice))
	for i, d := range dice {
		parts[i] = strconv.Itoa(d)
	}
	return strings.Join(parts, " ")
}
