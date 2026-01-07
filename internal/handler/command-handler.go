package handler

import (
	"context"
	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands"
	"log"
	"strings"
)

func (h *Handler) HandleCommand(quit chan bool, event *model.WebSocketEvent) error {
	cmds := commands.NewCommands(h.Settings, h.Mm, h.Cache, h.Pubsub)
	channelId := event.GetBroadcast().ChannelId
	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	sender := event.GetData()["sender_name"].(string)

	bc, err := cmds.NewBotCommand(post.Message, sender)
	if err != nil {
		return h.SendErrorResponse(post, err.Error())
	}
	bc.ResponseChannel = h.ResponseChannel
	if bc.ReplyChannel == nil || bc.ReplyChannel.Id == "" {
		bc.ReplyChannel, _, err = h.Mm.Client.GetChannel(context.Background(), channelId, "")
		if err != nil {
			return h.SendErrorResponse(post, err.Error())
		}
	}
	bc.Quit = quit

	err = cmds.CallCommand(bc)
	if err != nil {
		log.Printf("Error Executing command: %v", err)
		return h.SendErrorResponse(post, err.Error())
	}

	return err
}
