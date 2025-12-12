package games

import (
	"fmt"

	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

// InitFarkleGame constructs a FarkleGame ready for play without cache or bot wiring.
// You must provide at least two players; goal of 0 uses the default.
func InitFarkleGame(players []users.User, goal int) (FarkleGame, error) {
	if len(players) < 2 {
		return FarkleGame{}, fmt.Errorf("need at least 2 players")
	}
	if goal <= 0 {
		goal = defaultGoal
	}

	scores := make(map[string]int, len(players))
	for _, p := range players {
		scores[playerKey(p)] = 0
	}

	return FarkleGame{
		Players:       players,
		Scores:        scores,
		State:         statePlaying,
		CurrentTurn:   0,
		TurnPoints:    0,
		DiceRemaining: 6,
		LastRoll:      nil,
		TargetScore:   goal,
		FinalStart:    -1,
		ChannelId:     "local",
	}, nil
}

// FarkleRollTurn performs a roll for the provided player on the given game.
func FarkleRollTurn(game *FarkleGame, player users.User) (string, string, error) {
	return handleFarkleRoll(game, player)
}

// FarkleKeepDice applies the keep selection for the provided player.
func FarkleKeepDice(game *FarkleGame, player users.User, args []string) (string, error) {
	return handleFarkleKeep(game, player, args)
}

// FarkleBankTurn banks points for the provided player.
func FarkleBankTurn(game *FarkleGame, player users.User) (string, string, error) {
	return handleFarkleBank(game, player)
}
