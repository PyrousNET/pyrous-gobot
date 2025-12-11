package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/handler/games"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const botName = "Computer"

func main() {
	human, goal, botCount := parseConfig()

	players := []users.User{{Name: human, Id: human}}
	for i := 1; i <= botCount; i++ {
		suffix := ""
		if botCount > 1 {
			suffix = fmt.Sprintf(" %d", i)
		}
		players = append(players, users.User{Name: botName + suffix})
	}

	game, err := games.InitFarkleGame(players, goal)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Farkle local test. Players: %s. Goal: %d.\n", playerNames(players), game.TargetScore)
	fmt.Println("Commands: roll | keep <dice or N> | bank | quit")

	reader := bufio.NewReader(os.Stdin)
	for {
		current := game.Players[game.CurrentTurn]
		if strings.HasPrefix(current.Name, botName) {
			fmt.Println("\n--- Bot turn:", current.Name, "---")
			if done := botTurn(&game); done {
				return
			}
			time.Sleep(300 * time.Millisecond)
			continue
		}

		fmt.Printf("\n--- %s turn ---\n", current.Name)
		fmt.Printf("[%s] > ", current.Name)
		line, _ := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])
		args := parts[1:]

		switch cmd {
		case "roll":
			msg, endMsg, err := games.FarkleRollTurn(&game, current)
			printResult(msg, endMsg, err)
			if endMsg != "" {
				return
			}
		case "keep":
			msg, err := games.FarkleKeepDice(&game, current, args)
			printResult(msg, "", err)
		case "bank":
			msg, endMsg, err := games.FarkleBankTurn(&game, current)
			printResult(msg, endMsg, err)
			if endMsg != "" {
				return
			}
		case "quit":
			fmt.Println("Exiting.")
			return
		default:
			fmt.Println("Unknown command. Try: roll | keep <dice or N> | bank | quit")
		}
	}
}

func botTurn(game *games.FarkleGame) bool {
	player := game.Players[game.CurrentTurn]
	fmt.Printf("[%s] rolling...\n", player.Name)
	msg, endMsg, err := games.FarkleRollTurn(game, player)
	printResult(msg, endMsg, err)
	if err != nil || endMsg != "" {
		return endMsg != ""
	}

	if len(game.LastRoll) > 0 && gamesHasScoring(game.LastRoll) {
		keepArgs := []string{strconv.Itoa(game.DiceRemaining), "dice"}
		msg, err = games.FarkleKeepDice(game, player, keepArgs)
		printResult(msg, "", err)
		if err != nil {
			return false
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Simple policy: bank if turn points >= 750 or dice remaining <= 2 and turn points > 0
	if game.TurnPoints >= 750 || (game.DiceRemaining <= 2 && game.TurnPoints > 0) {
		msg, endMsg, err = games.FarkleBankTurn(game, player)
		printResult(msg, endMsg, err)
		if endMsg != "" {
			return true
		}
	} else {
		fmt.Printf("[%s] rolling again...\n", player.Name)
	}
	return false
}

func gamesHasScoring(roll []int) bool {
	return games.HasScoringDiceExported(roll)
}

func printResult(msg, endMsg string, err error) {
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if msg != "" {
		fmt.Println(msg)
	}
	if endMsg != "" {
		fmt.Println(endMsg)
	}
}

func parseGoal() int {
	if v := os.Getenv("FARKLE_GOAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 0
}

func parseBotCount() int {
	if v := os.Getenv("FARKLE_BOTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

func parseConfig() (player string, goal int, bots int) {
	// Defaults from env
	player = os.Getenv("FARKLE_PLAYER")
	if player == "" {
		player = "You"
	}
	goal = parseGoal()
	bots = parseBotCount()

	// Flags override env
	playerFlag := flag.String("player", player, "Human player name")
	goalFlag := flag.Int("goal", goal, "Target score to win")
	botsFlag := flag.Int("bots", bots, "Number of computer opponents")
	flag.Parse()

	if *playerFlag != "" {
		player = *playerFlag
	}
	if *goalFlag > 0 {
		goal = *goalFlag
	}
	if *botsFlag > 0 {
		bots = *botsFlag
	}
	return
}

func playerNames(players []users.User) string {
	names := make([]string, 0, len(players))
	for _, p := range players {
		names = append(names, p.Name)
	}
	return strings.Join(names, ", ")
}
