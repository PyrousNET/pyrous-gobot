package tictactoe

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	bg "github.com/quibbble/go-boardgame"
	"github.com/quibbble/go-boardgame/pkg/bgerr"
)

const (
	minTeams = 2
	maxTeams = 2
)

// TicTacToe implements the bg.BoardGame interface
type TicTacToe struct {
	State   *State
	actions []*bg.BoardGameAction
}

// NewTicTacToe creates a new game instance
func NewTicTacToe(options *bg.BoardGameOptions) (*TicTacToe, error) {
	if len(options.Teams) < minTeams {
		return nil, &bgerr.Error{
			Err:    fmt.Errorf("at least %d players required to create a game of %s", minTeams, key),
			Status: bgerr.StatusTooFewTeams,
		}
	} else if len(options.Teams) > maxTeams {
		return nil, &bgerr.Error{
			Err:    fmt.Errorf("at most %d players allowed to create a game of %s", maxTeams, key),
			Status: bgerr.StatusTooManyTeams,
		}
	}
	return &TicTacToe{
		State:   newState(options.Teams),
		actions: make([]*bg.BoardGameAction, 0),
	}, nil
}

// Do performs an action on the game
func (t *TicTacToe) Do(action *bg.BoardGameAction) error {
	if len(t.State.Winners) > 0 {
		return &bgerr.Error{
			Err:    fmt.Errorf("game already over"),
			Status: bgerr.StatusGameOver,
		}
	}
	switch action.ActionType {
	case ActionMarkLocation:
		var details MarkLocationActionDetails
		if err := mapstructure.Decode(action.MoreDetails, &details); err != nil {
			return &bgerr.Error{
				Err:    err,
				Status: bgerr.StatusInvalidActionDetails,
			}
		}
		if err := t.State.MarkLocation(action.Team, details.Row, details.Column); err != nil {
			return err
		}
		t.actions = append(t.actions, action)
	default:
		return &bgerr.Error{
			Err:    fmt.Errorf("cannot process action type %s", action.ActionType),
			Status: bgerr.StatusUnknownActionType,
		}
	}
	return nil
}

// GetSnapshot returns the game from team's view
// TicTacToe does not require hiding information from different players so everything can be returned
func (t *TicTacToe) GetSnapshot(team ...string) (*bg.BoardGameSnapshot, error) {
	if len(team) > 1 {
		return nil, &bgerr.Error{
			Err:    fmt.Errorf("get snapshot requires zero or one team"),
			Status: bgerr.StatusTooManyTeams,
		}
	}
	var targets []*bg.BoardGameAction
	if len(t.State.Winners) == 0 && (len(team) == 0 || (len(team) == 1 && team[0] == t.State.Turn)) {
		targets = t.State.targets()
	}
	return &bg.BoardGameSnapshot{
		Turn:    t.State.Turn,
		Teams:   t.State.Teams,
		Winners: t.State.Winners,
		MoreData: TicTacToeSnapshotData{
			Board: t.State.Board,
		},
		Targets: targets,
		Actions: t.actions,
		Message: t.State.message(),
	}, nil
}
