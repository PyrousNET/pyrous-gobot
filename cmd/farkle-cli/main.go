package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/handler/games"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const botName = "Computer"

func main() {
	human := os.Getenv("FARKLE_PLAYER")
	if human == "" {
		human = "You"
	}
	goal := parseGoal()

	game, err := games.InitFarkleGame([]users.User{
		{Name: human, Id: human},
		{Name: botName},
	}, goal)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Farkle local test. Players: %s vs %s. Goal: %d.\n", human, botName, game.TargetScore)
	fmt.Println("Commands: roll | keep <dice or N> | bank | quit")

	reader := bufio.NewReader(os.Stdin)
	for {
		current := game.Players[game.CurrentTurn]
		if current.Name == botName {
			if done := botTurn(&game); done {
				return
			}
			continue
		}

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
	fmt.Println("[Bot] rolling...")
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
	}

	// Simple policy: bank if turn points >= 750 or dice remaining <= 2 and turn points > 0
	if game.TurnPoints >= 750 || (game.DiceRemaining <= 2 && game.TurnPoints > 0) {
		msg, endMsg, err = games.FarkleBankTurn(game, player)
		printResult(msg, endMsg, err)
		if endMsg != "" {
			return true
		}
	} else {
		fmt.Println("[Bot] rolling again...")
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
