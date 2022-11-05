package games

import (
	"fmt"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games/rps"
	"strings"

	"github.com/google/uuid"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

func (bg BotGame) Rps(event BotGame) error {
	response := comms.Response{
		Type:           "command",
		ReplyChannelId: event.ReplyChannel.Id,
	}
	var choice, channel string
	fmt.Sscanf(event.body, "%s %s", &channel, &choice)
	foundChannel, cErr := event.mm.GetChannelByName(channel)
	if cErr == nil && foundChannel != nil {
		response.ReplyChannelId = foundChannel.Id
	} else {
		foundChannel = event.ReplyChannel
	}
	playerUser, _, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.Cache)
	response.UserId = playerUser.Id
	if err != nil {
		return err
	}
	player, perr := rps.GetPlayer(playerUser, foundChannel.Id, event.Cache)
	rpsEvent := rps.RpsBotGame{
		ReplyChannel:    event.ReplyChannel,
		ResponseChannel: event.ResponseChannel,
		Cache:           event.Cache,
	}
	opponent, oErr := rps.FindApponent(rpsEvent, player, foundChannel.Id)
	if perr != nil || !rps.Playing(player) {
		if oErr == nil && rps.Playing(opponent) {
			channelId, ok, _ := event.Cache.Get(opponent.RpsPlaying)
			if ok && event.ReplyChannel != nil && channelId == foundChannel.Id {
				player.RpsPlaying = opponent.RpsPlaying
				response.Type = "dm"
				response.UserId = playerUser.Id
				response.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps %s rock)", event.ReplyChannel.Name)
			}
		} else {
			id, e := uuid.NewRandom()
			dmResponse := comms.Response{
				Type:   "dm",
				UserId: playerUser.Id,
			}
			event.Cache.Put(id.String(), event.ReplyChannel.Id)
			response.Message = fmt.Sprintf("/echo %s is looking for an opponent in RPS.", event.sender)
			dmResponse.Message = fmt.Sprintf("Would you like to throw Rock, Paper or Scissors (Usage: $rps %s rock)", event.ReplyChannel.Name)
			event.ResponseChannel <- dmResponse
			if e != nil {
				return e
			}
			player.RpsPlaying = id.String()
		}
	}

	if event.body != "" {
		switch strings.ToLower(choice) {
		case "rock", "paper", "scissors":
			player.Rps = strings.ToLower(choice)
			response.Type = "dm"
			response.UserId = playerUser.Id
			response.Message = fmt.Sprintf("I have you down for: %s", strings.Title(strings.ToLower(choice)))
		default:
			response.Type = "dm"
			response.UserId = playerUser.Id
			response.Message = fmt.Sprintf(`Uh, %s isn't an option. Try {channel} rock, paper or scissors'`, choice)
		}
	}

	if oErr == nil && opponent.Name != "" {
		winners, hasWinner := rps.GetWinner(player, opponent)
		if hasWinner {
			channelId, ok, _ := event.Cache.Get(player.RpsPlaying)
			response.Type = "command"
			if ok {
				response.ReplyChannelId = channelId.(string)
				if len(winners) > 1 {
					response.Message = fmt.Sprintf("/echo The RPS game between %s and %s ended in a draw.", player.Name, opponent.Name)
				} else {
					response.Message = fmt.Sprintf("/echo The RPS game between %s and %s ended with %s winning.", player.Name, opponent.Name, winners[0].Name)
				}
			}

			rps.DeleteGame(player.RpsPlaying, event.Cache)
			rps.DeleteRps(player, foundChannel.Id, event.Cache)
			rps.DeleteRps(opponent, foundChannel.Id, event.Cache)
			event.ResponseChannel <- response
			return err
		}

		rps.UpdateRps(opponent, foundChannel.Id, event.Cache)
	}

	rps.UpdateRps(player, foundChannel.Id, event.Cache)

	event.ResponseChannel <- response

	return err
}
