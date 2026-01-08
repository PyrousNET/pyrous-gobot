package handler

import (
	"context"
	"log"
	"strings"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games"
)

func (h *Handler) HandleGame(quit chan bool, event *model.WebSocketEvent) error {
	gms := games.NewGames(h.Settings, h.Mm, h.Cache)
	channelId := event.GetBroadcast().ChannelId
	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	sender := event.GetData()["sender_name"].(string)
	var e error
	if channelId == "" {
		channelId = post.ChannelId
	}

	bg, err := gms.NewBotGame(post.Message, sender)
	bg.ResponseChannel = h.ResponseChannel // They are shared now!
	if err != nil {
		return h.SendErrorResponse(post, err.Error())
	}
	bg.ReplyChannel, _, e = h.Mm.Client.GetChannel(context.Background(), channelId, "")
	if e != nil {
		return h.SendErrorResponse(post, e.Error())
	}

	err = gms.CallGame(bg)
	if err != nil {
		log.Printf("error executing game: %v", err)
		return h.SendErrorResponse(post, err.Error())
	}
	return err
}
