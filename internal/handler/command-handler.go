package handler

import (
	"log"
	"strings"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands"
)

func (h *Handler) HandleCommand(quit chan bool, event *model.WebSocketEvent) error {
	cmds := commands.NewCommands(h.Settings, h.Mm, h.Cache)
	channelId := event.GetBroadcast().ChannelId
	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	sender := event.GetData()["sender_name"].(string)

	bc, err := cmds.NewBotCommand(post.Message, sender)
	if err != nil {
		return h.SendErrorResponse(post, err.Error())
	}

	r, err := cmds.CallCommand(bc)
	if err != nil {
		log.Printf("Error Executing command: %v", err)
		return h.SendErrorResponse(post, err.Error())
	}

	if r.Channel == "" && bc.ReplyChannel.Id == "" {
		r.Channel = event.GetBroadcast().ChannelId
	} else {
		r.Channel = bc.ReplyChannel.Id
		r.Type = "command"
		checkMsg := strings.Split(r.Message, " ")
		if checkMsg[0] != "/echo" {
			r.Message = "/echo " + r.Message
		}
	}

	if "" == r.Channel {
		r.Channel = channelId
	}

	if r.Type != "shutdown" {
		dmchannel, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
		if r.Channel == dmchannel.Id {
			r.Type = "dm"
		}
	}

	if r.Message != "" {
		switch r.Type {
		case "post":
			err = h.Mm.SendMsgToChannel(r.Message, r.Channel, post)
			if r.Message2 != "" {
				time.Sleep(r.Delay)
				err = h.Mm.SendMsgToChannel(r.Message2, r.Channel, post)
			}
		case "command":
			err = h.Mm.SendCmdToChannel(r.Message, r.Channel, post)
		case "dm":
			c, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = r.Message

			_, _, err = h.Mm.Client.CreatePost(replyPost)
			if r.Message2 != "" {
				time.Sleep(r.Delay)
				replyPost := &model.Post{}
				replyPost.ChannelId = c.Id
				replyPost.Message = r.Message2

				_, _, err = h.Mm.Client.CreatePost(replyPost)
			}
		case "shutdown":
			c, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = r.Message

			_, _, err := h.Mm.Client.CreatePost(replyPost)

			err = h.Mm.SendMsgToChannel("Awe, Crap!", r.Channel, post)
			if err != nil {
				log.Print(err)
			}

			h.Cache.Put("restart-usr", post.UserId)

			quit <- true
		}
	}

	return err
}
