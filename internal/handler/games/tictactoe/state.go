package tictactoe

import (
	"fmt"
	"strings"

	bg "github.com/quibbble/go-boardgame"
	"github.com/quibbble/go-boardgame/pkg/bgerr"
)

const size = 3

var (
	indexToMark = map[int]string{0: "O", 1: "X"}
	markToIndex = map[string]int{"O": 0, "X": 1}
)

// state handles all the internal game logic for the game
type State struct {
	Turn    string
	Teams   []string
	Winners []string
	Board   [size][size]string
}

func newState(teams []string) *State {
	return &State{
		Turn:    teams[0],
		Teams:   teams,
		Winners: make([]string, 0),
		Board:   [size][size]string{},
	}
}

func (s *State) MarkLocation(team string, row, column int) error {
	index := indexOf(s.Teams, team)
	if index < 0 {
		return &bgerr.Error{
			Err:    fmt.Errorf("%s not playing the game", team),
			Status: bgerr.StatusInvalidActionDetails,
		}
	}
	if team != s.Turn {
		return &bgerr.Error{
			Err:    fmt.Errorf("%s cannot play on %s turn", team, s.Turn),
			Status: bgerr.StatusInvalidAction,
		}
	}
	if row < 0 || row >= size || column < 0 || column >= size {
		return &bgerr.Error{
			Err:    fmt.Errorf("row or column out of bounds"),
			Status: bgerr.StatusInvalidActionDetails,
		}
	}
	if s.Board[row][column] != "" {
		return &bgerr.Error{
			Err:    fmt.Errorf("%d,%d already marked", row, column),
			Status: bgerr.StatusInvalidAction,
		}
	}

	// mark index
	s.Board[row][column] = indexToMark[index]

	// check and update winner
	if winner(s.Board) != "" {
		s.Winners = []string{s.Teams[markToIndex[winner(s.Board)]]}
	} else if draw(s.Board) {
		s.Winners = s.Teams
	}

	// update turn
	s.Turn = s.Teams[(index+1)%2]
	return nil
}

func (s *State) targets() []*bg.BoardGameAction {
	targets := make([]*bg.BoardGameAction, 0)
	for r, row := range s.Board {
		for c, loc := range row {
			if loc == "" {
				targets = append(targets, &bg.BoardGameAction{
					Team:       s.Turn,
					ActionType: ActionMarkLocation,
					MoreDetails: MarkLocationActionDetails{
						Row:    r,
						Column: c,
					},
				})
			}
		}
	}
	return targets
}

func (s *State) message() string {
	message := fmt.Sprintf("%s must mark a location", s.Turn)
	if len(s.Winners) > 0 {
		message = fmt.Sprintf("%s tie", strings.Join(s.Winners, " and "))
		if len(s.Winners) == 1 {
			message = fmt.Sprintf("%s wins", s.Winners[0])
		}
	}
	return message
}

func winner(board [size][size]string) string {
	for i := 0; i < size; i++ {
		// check rows
		if board[i][0] != "" && board[i][0] == board[i][1] && board[i][0] == board[i][2] {
			return board[i][0]
		}
		// check columns
		if board[0][i] != "" && board[0][i] == board[1][i] && board[0][i] == board[2][i] {
			return board[0][i]
		}
	}
	// check diagonal
	if board[0][0] != "" && board[0][0] == board[1][1] && board[0][0] == board[2][2] {
		return board[0][0]
	}
	// check diagonal
	if board[2][0] != "" && board[2][0] == board[1][1] && board[2][0] == board[0][2] {
		return board[2][0]
	}
	return ""
}

func draw(board [size][size]string) bool {
	for _, row := range board {
		for _, loc := range row {
			if loc == "" {
				return false
			}
		}
	}
	return true
}

func indexOf(items []string, item string) int {
	for i, it := range items {
		if it == item {
			return i
		}
	}
	return -1
}
