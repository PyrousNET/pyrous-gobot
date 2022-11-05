package handler

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands"
	"log"
	"strings"
)

func (h *Handler) HandleCommand(quit chan bool, event *model.WebSocketEvent) error {
	cmds := commands.NewCommands(h.Settings, h.Mm, h.Cache)
	channelId := event.GetBroadcast().ChannelId
	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	sender := event.GetData()["sender_name"].(string)

	bc, err := cmds.NewBotCommand(post.Message, sender)
	bc.ResponseChannel = h.ResponseChannel
	bc.ReplyChannel, _, err = h.Mm.Client.GetChannel(channelId, "")
	bc.Quit = quit
	if err != nil {
		return h.SendErrorResponse(post, err.Error())
	}

	err = cmds.CallCommand(bc)
	if err != nil {
		log.Printf("Error Executing command: %v", err)
		return h.SendErrorResponse(post, err.Error())
	}

	return err
}
