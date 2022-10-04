package games

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	ttt "github.com/pyrousnet/pyrous-gobot/internal/handler/games/tictactoe"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
	boardgame "github.com/quibbble/go-boardgame"
)

type GameData struct {
	Players []users.User  `json:"players,omitempty"`
	Game    ttt.TicTacToe `json:"game"`
}

type TicTacToeGame struct {
	Data    GameData
	Channel *model.Channel
	Cache   cache.Cache
	Player  string
}

func (h BotGameHelp) Ttt(request BotGame) (response HelpResponse) {
	response.Help = "Play Tic Tac Toe in the current channel."
	response.Description = "Play Tic Tac Toe in the current channel."

	return response
}

func (bg BotGame) Ttt(event BotGame) (response Response, err error) {
	g, err := NewTicTacToeGame(event)
	if err != nil {
		log.Println("It's seeing the error")
		switch err.Error() {
		case "please wait for player 2 to join":
			log.Println("only one player")
			response.Type = "post"
			response.Message = fmt.Sprintf("%s is looking for an opponent to play Tic Tac Toe.", g.Player)

			return response, nil
		case "maximum number of players":
			response.Type = "post"
			response.Message = fmt.Sprintf(
				"Woah there, pardner! %s and %s are already playing a game in %s!",
				g.Data.Players[0].Name,
				g.Data.Players[1].Name,
				g.Channel.Name,
			)

			return response, nil
		case "ready to start":
			response.Type = "post"
			response.Message = g.PrintBoard()

			return response, nil
		default:
			response.Type = "dm"
			response.Message = fmt.Sprintf("There was an error starting your game in %s: %v", g.Channel.Name, err)

			return response, err
		}
	}

	if event.body != "" {
		var row, column int
		_, err = fmt.Sscanf(event.body, "%v%v", &row, &column)
		if err != nil {
			response.Type = "dm"
			response.Message = fmt.Sprintf("I didn't understand your command: %v\n%s", err, g.PrintHelp())

			return response, err
		}

		err = g.Data.Game.Do(&boardgame.BoardGameAction{
			Team:       strings.TrimLeft(event.sender, "@"),
			ActionType: "MarkLocation",
			MoreDetails: ttt.MarkLocationActionDetails{
				Row:    row - 1,
				Column: column - 1,
			},
		})
		if err != nil {
			response.Type = "dm"
			response.Message = fmt.Sprintf("invalid move: %v\n%s", err, g.PrintHelp())

			return response, err
		}
		err = g.CacheGameData()
		if err != nil {
			response.Type = "dm"
			response.Message = fmt.Sprintf("could not cache game data: %v", err)
		}

		response.Type = "command"
		response.Message = fmt.Sprintf("/echo %s", g.PrintBoard())

		return response, err
	}

	response.Type = "post"
	response.Message = g.PrintBoard()

	return response, err
}

func NewTicTacToeGame(event BotGame) (*TicTacToeGame, error) {
	g := &TicTacToeGame{
		Channel: event.ReplyChannel,
		Cache:   event.cache,
		Player:  event.sender,
	}

	player, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil {
		return g, fmt.Errorf("user not found: %v", err)
	}

	gd, ok, err := g.GetGameData()
	if err != nil {
		return g, fmt.Errorf("error fetching game data: %v", err)
	}

	if !ok {
		// It's a new game. Make new GameData and let player 1 know to wait for player 2
		g.Data = GameData{}
		g.Data.Players = append(g.Data.Players, player)
		err = g.CacheGameData()
		if err != nil {
			return g, fmt.Errorf("could not cache game data: %v", err)
		}

		return g, fmt.Errorf("please wait for player 2 to join")
	}

	g.Data = gd

	if len(g.Data.Players) == 1 {
		// Player 2 has joined!
		g.Data.Players = append(g.Data.Players, player)
		game, err := ttt.NewTicTacToe(
			&boardgame.BoardGameOptions{
				Teams: []string{
					g.Data.Players[0].Name,
					g.Data.Players[1].Name,
				},
			},
		)
		if err != nil {
			log.Print(err)
			return g, fmt.Errorf("could not create game: %v", err)
		}

		g.Data.Game = *game

		err = g.CacheGameData()
		if err != nil {
			return g, fmt.Errorf("could not cache game data: %v", err)
		}

		return g, fmt.Errorf("ready to start")
	}

	if len(g.Data.Players) == 2 && player.Id != g.Data.Players[0].Id && player.Id != g.Data.Players[1].Id {
		return g, fmt.Errorf("maximum number of players")
	}

	return g, nil
}

func (tttg *TicTacToeGame) GetGameData() (GameData, bool, error) {
	var gd GameData
	key := fmt.Sprintf("tttg_%s", tttg.Channel.Id)

	r, ok, err := tttg.Cache.Get(key)
	if ok {
		if reflect.TypeOf(r).String() != "[]uint8" {
			json.Unmarshal([]byte(r.(string)), &gd)
		} else {
			json.Unmarshal(r.([]byte), &gd)
		}
		return gd, ok, nil
	}
	return gd, ok, err

}

func (tttg *TicTacToeGame) CacheGameData() error {
	key := fmt.Sprintf("tttg_%s", tttg.Channel.Id)
	dj, err := json.Marshal(tttg.Data)
	if err != nil {
		return fmt.Errorf("could not save game data: %v", err)
	}
	tttg.Cache.Put(key, dj)

	return nil
}

func (tttg *TicTacToeGame) DeleteGameData() {
	key := fmt.Sprintf("tttg_%s", tttg.Channel.Id)
	tttg.Cache.Clean(key)
}

func (tttg *TicTacToeGame) PrintBoard() string {
	snapshot, _ := tttg.Data.Game.GetSnapshot(tttg.Player)
	sb := snapshot.MoreData.(ttt.TicTacToeSnapshotData).Board

	snap := fmt.Sprintf("||1|2|3|\n|:-:|:-:|:-:|:-:|\n")
	for r, row := range sb {
		snap = fmt.Sprintf("%s|%v|", snap, r+1)
		for _, c := range row {
			snap = fmt.Sprintf("%s%s|", snap, c)
		}
		snap = fmt.Sprintf("%s \n", snap)

	}

	switch len(snapshot.Winners) {
	case 1:
		snap = fmt.Sprintf(
			"%s\n\nThe Tic Tac Toe game between %s and %s is now over with %s winning!\n",
			snap,
			tttg.Data.Players[0].Name,
			tttg.Data.Players[1].Name,
			snapshot.Winners[0],
		)
		tttg.DeleteGameData()
	case 2:
		snap = fmt.Sprintf(
			"%s\n\nThe Tic Tac Toe game between %s and %s is now over and ended in a draw.\n",
			snap,
			tttg.Data.Players[0].Name,
			tttg.Data.Players[1].Name,
		)
		tttg.DeleteGameData()
	default:
		snap = fmt.Sprintf("%s\n\n%s make your mark!\n%s", snap, snapshot.Turn, tttg.PrintHelp())
	}

	return snap
}

func (tttg *TicTacToeGame) PrintHelp() string {
	return "Help: `$ttt {row} {column}` E.G.: `$ttt 1 2`"
}
