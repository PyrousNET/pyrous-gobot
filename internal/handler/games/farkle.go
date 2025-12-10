package games

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const (
	farklePrefix    = "farkle_"
	defaultGoal     = 5000
	stateLobby      = "lobby"
	statePlaying    = "playing"
	stateFinalRound = "final"
)

type FarkleGame struct {
	Players       []users.User   `json:"players"`
	Scores        map[string]int `json:"scores"`
	State         string         `json:"state"`
	CurrentTurn   int            `json:"current_turn"`
	TurnPoints    int            `json:"turn_points"`
	DiceRemaining int            `json:"dice_remaining"`
	LastRoll      []int          `json:"last_roll"`
	TargetScore   int            `json:"target_score"`
	FinalStart    int            `json:"final_start"`
	ChannelId     string         `json:"channel_id"`
}

var rollDiceFn = rollDice

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (bg BotGame) Farkle(event BotGame) error {
	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	if err != nil {
		return err
	}

	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Type:           "command",
		UserId:         player.Id,
	}

	game, ok, err := loadFarkle(event.ReplyChannel.Id, event.Cache)
	if err != nil {
		return err
	}
	if !ok {
		game = newFarkleGame(event.ReplyChannel.Id, player)
		saveFarkle(game, event.Cache)
		response.Message = fmt.Sprintf("/echo %s started a Farkle lobby. Use `$farkle start` when ready. Players: %s", player.Name, formatPlayerNames(game.Players))
		event.ResponseChannel <- response
		return nil
	}

	parts := strings.Fields(event.body)
	if len(parts) == 0 {
		// Join lobby or show status.
		if game.State == stateLobby {
			if addPlayerToFarkle(&game, player) {
				saveFarkle(game, event.Cache)
				response.Message = fmt.Sprintf("/echo %s joined the Farkle lobby. Players: %s", player.Name, formatPlayerNames(game.Players))
			} else {
				response.Message = fmt.Sprintf("/echo Farkle lobby players: %s. Use `$farkle start` when ready.", formatPlayerNames(game.Players))
			}
		} else {
			response.Message = fmt.Sprintf("/echo Game in progress. %s to play. %s", game.Players[game.CurrentTurn].Name, formatScores(game))
		}
		event.ResponseChannel <- response
		return nil
	}

	directive := strings.ToLower(parts[0])
	args := parts[1:]

	switch directive {
	case "start":
		msg, err := handleFarkleStart(&game, player)
		if err != nil {
			response.Message = fmt.Sprintf("/echo %v", err)
		} else {
			saveFarkle(game, event.Cache)
			response.Message = msg
		}
	case "roll":
		msg, endMsg, err := handleFarkleRoll(&game, player)
		if err != nil {
			response.Message = fmt.Sprintf("/echo %v", err)
		} else {
			saveFarkle(game, event.Cache)
			if endMsg != "" {
				response.Message = fmt.Sprintf("/echo %s\n%s", msg, endMsg)
				clearFarkle(event.ReplyChannel.Id, event.Cache)
			} else {
				response.Message = msg
			}
		}
	case "keep":
		msg, err := handleFarkleKeep(&game, player, args)
		if err != nil {
			response.Message = fmt.Sprintf("/echo %v", err)
		} else {
			saveFarkle(game, event.Cache)
			response.Message = msg
		}
	case "bank":
		msg, endMsg, err := handleFarkleBank(&game, player)
		if err != nil {
			response.Message = fmt.Sprintf("/echo %v", err)
		} else {
			saveFarkle(game, event.Cache)
			if endMsg != "" {
				response.Message = fmt.Sprintf("/echo %s\n%s", msg, endMsg)
				// Game ended; clear state.
				clearFarkle(event.ReplyChannel.Id, event.Cache)
			} else {
				response.Message = msg
			}
		}
	case "quit", "leave", "end":
		clearFarkle(event.ReplyChannel.Id, event.Cache)
		response.Message = "/echo Farkle game cleared."
	default:
		response.Message = "/echo Unknown farkle command. Try `$farkle start`, `$farkle roll`, `$farkle keep <dice>`, `$farkle bank`."
	}

	event.ResponseChannel <- response
	return nil
}

func newFarkleGame(channelId string, starter users.User) FarkleGame {
	key := playerKey(starter)
	return FarkleGame{
		Players:       []users.User{starter},
		Scores:        map[string]int{key: 0},
		State:         stateLobby,
		CurrentTurn:   0,
		TurnPoints:    0,
		DiceRemaining: 6,
		TargetScore:   defaultGoal,
		FinalStart:    -1,
		ChannelId:     channelId,
	}
}

func addPlayerToFarkle(game *FarkleGame, player users.User) bool {
	key := playerKey(player)
	for _, p := range game.Players {
		if playerKey(p) == key {
			return false
		}
	}
	game.Players = append(game.Players, player)
	game.Scores[key] = 0
	return true
}

func loadFarkle(channelId string, c cache.Cache) (FarkleGame, bool, error) {
	var game FarkleGame
	r, ok, err := c.Get(farklePrefix + channelId)
	if err != nil || !ok {
		return game, ok, err
	}

	switch v := r.(type) {
	case []byte:
		err = json.Unmarshal(v, &game)
	case string:
		err = json.Unmarshal([]byte(v), &game)
	default:
		return game, false, fmt.Errorf("unexpected farkle cache type")
	}
	return game, true, err
}

func saveFarkle(game FarkleGame, c cache.Cache) {
	data, _ := json.Marshal(game)
	c.Put(farklePrefix+game.ChannelId, data)
}

func clearFarkle(channelId string, c cache.Cache) {
	c.Clean(farklePrefix + channelId)
}

func handleFarkleStart(game *FarkleGame, player users.User) (string, error) {
	if game.State != stateLobby {
		return "", fmt.Errorf("game already started")
	}
	if len(game.Players) < 2 {
		return "", fmt.Errorf("need at least 2 players to start")
	}
	if !isPlayerIn(game.Players, player) {
		return "", fmt.Errorf("join the lobby first with `$farkle`")
	}
	game.State = statePlaying
	game.CurrentTurn = 0
	game.TurnPoints = 0
	game.DiceRemaining = 6
	game.LastRoll = nil
	return fmt.Sprintf("/echo Farkle started! Goal: %d. Turn order: %s. %s begins—use `$farkle roll`.", game.TargetScore, formatPlayerNames(game.Players), game.Players[0].Name), nil
}

func handleFarkleRoll(game *FarkleGame, player users.User) (string, string, error) {
	if err := ensurePlayerTurn(game, player); err != nil {
		return "", "", err
	}
	if len(game.LastRoll) > 0 {
		return "", "", fmt.Errorf("you must keep or bank before rolling again")
	}

	roll := rollDiceFn(game.DiceRemaining)
	game.LastRoll = roll

	if !hasScoringDice(roll) {
		// Farkle
		next := advanceTurn(game)
		msg := fmt.Sprintf("/echo %s rolled %v and farkled. Turn points lost. %s to play. %s", player.Name, roll, game.Players[next].Name, formatScores(*game))
		if game.State == stateFinalRound && next == game.FinalStart {
			winner := determineWinner(game)
			endMsg := fmt.Sprintf("Game over! Winner: %s with %d points. Final scores: %s", winner.Name, game.Scores[playerKey(winner)], formatScores(*game))
			return msg, endMsg, nil
		}
		return msg, "", nil
	}

	return fmt.Sprintf("/echo %s rolled %v. Turn points: %d. Select scoring dice with `$farkle keep <dice>` or bank with `$farkle bank`.", player.Name, roll, game.TurnPoints), "", nil
}

func handleFarkleKeep(game *FarkleGame, player users.User, args []string) (string, error) {
	if err := ensurePlayerTurn(game, player); err != nil {
		return "", err
	}
	if len(game.LastRoll) == 0 {
		return "", fmt.Errorf("roll first with `$farkle roll`")
	}

	selection, warn, err := selectDiceToKeep(args, game.LastRoll)
	if err != nil {
		return "", err
	}
	if !isSubset(selection, game.LastRoll) {
		return "", fmt.Errorf("kept dice must come from the last roll %v", game.LastRoll)
	}

	score, err := scoreSelection(selection)
	if err != nil {
		return "", err
	}

	game.TurnPoints += score
	game.DiceRemaining -= len(selection)
	if game.DiceRemaining == 0 {
		game.DiceRemaining = 6 // hot dice
	}
	game.LastRoll = nil

	msg := fmt.Sprintf("/echo Kept %v for %d points. Turn points: %d. Dice remaining: %d. Roll again with `$farkle roll` or bank with `$farkle bank`.", selection, score, game.TurnPoints, game.DiceRemaining)
	if warn != "" {
		msg = fmt.Sprintf("/echo %s\n%s", warn, strings.TrimPrefix(msg, "/echo "))
	}

	return msg, nil
}

func handleFarkleBank(game *FarkleGame, player users.User) (string, string, error) {
	if err := ensurePlayerTurn(game, player); err != nil {
		return "", "", err
	}

	key := playerKey(player)
	total := game.Scores[key] + game.TurnPoints
	game.Scores[key] = total
	game.TurnPoints = 0
	game.LastRoll = nil
	game.DiceRemaining = 6

	msg := fmt.Sprintf("/echo %s banked. Total: %d. %s", player.Name, total, formatScores(*game))

	endMsg := ""
	if game.State != stateFinalRound && total >= game.TargetScore {
		game.State = stateFinalRound
		game.FinalStart = game.CurrentTurn
		msg = fmt.Sprintf("%s\nFinal round begins! Others get one last turn to beat %s.", msg, player.Name)
	}

	next := advanceTurn(game)
	msg = fmt.Sprintf("%s\n%s to play.", msg, game.Players[next].Name)

	if game.State == stateFinalRound && next == game.FinalStart {
		// Final round complete, determine winner.
		winner := determineWinner(game)
		endMsg = fmt.Sprintf("Game over! Winner: %s with %d points. Final scores: %s", winner.Name, game.Scores[playerKey(winner)], formatScores(*game))
	}

	return msg, endMsg, nil
}

func ensurePlayerTurn(game *FarkleGame, player users.User) error {
	if !isPlayerIn(game.Players, player) {
		return fmt.Errorf("you are not in this game")
	}
	if game.State == stateLobby {
		return fmt.Errorf("game has not started yet. Use `$farkle start`.")
	}
	if playerKey(game.Players[game.CurrentTurn]) != playerKey(player) {
		return fmt.Errorf("it's %s's turn", game.Players[game.CurrentTurn].Name)
	}
	return nil
}

func advanceTurn(game *FarkleGame) int {
	game.CurrentTurn = (game.CurrentTurn + 1) % len(game.Players)
	game.TurnPoints = 0
	game.DiceRemaining = 6
	game.LastRoll = nil
	return game.CurrentTurn
}

func isPlayerIn(players []users.User, u users.User) bool {
	key := playerKey(u)
	for _, p := range players {
		if playerKey(p) == key {
			return true
		}
	}
	return false
}

func formatPlayerNames(players []users.User) string {
	names := make([]string, 0, len(players))
	for _, p := range players {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}

func formatScores(game FarkleGame) string {
	type entry struct {
		name  string
		score int
	}
	var scores []entry
	for _, p := range game.Players {
		scores = append(scores, entry{name: p.Name, score: game.Scores[playerKey(p)]})
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].name < scores[j].name
		}
		return scores[i].score > scores[j].score
	})

	parts := make([]string, 0, len(scores))
	for _, s := range scores {
		parts = append(parts, fmt.Sprintf("%s: %d", s.name, s.score))
	}
	return "Scores - " + strings.Join(parts, " | ")
}

func rollDice(n int) []int {
	out := make([]int, n)
	for i := 0; i < n; i++ {
		out[i] = rand.Intn(6) + 1
	}
	return out
}

func determineWinner(game *FarkleGame) users.User {
	bestIdx := 0
	bestScore := -1
	for idx, p := range game.Players {
		if score := game.Scores[playerKey(p)]; score > bestScore {
			bestScore = score
			bestIdx = idx
		}
	}
	return game.Players[bestIdx]
}

func playerKey(u users.User) string {
	if u.Id != "" {
		return u.Id
	}
	return u.Name
}
