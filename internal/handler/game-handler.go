package handler

import (
	"fmt"
	"log"
	"strings"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/games"
)

func (h *Handler) HandleGame(quit chan bool, event *model.WebSocketEvent) error {
	gms := games.NewGames(h.Settings, h.Mm, h.Cache)
	channelId := event.GetBroadcast().ChannelId
	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	sender := event.GetData()["sender_name"].(string)
	var e error

	bg, err := gms.NewBotGame(post.Message, sender)
	if err != nil {
		return h.SendErrorResponse(post, err.Error())
	}
	bg.ReplyChannel, _, e = h.Mm.Client.GetChannel(channelId, "")
	if e != nil {
		return h.SendErrorResponse(post, e.Error())
	}

	r, err := gms.CallGame(bg)
	if err != nil {
		log.Printf("error executing game: %v", err)
		return h.SendErrorResponse(post, err.Error())
	}

	if r.Channel == "" && bg.ReplyChannel.Id == "" {
		r.Channel = event.GetBroadcast().ChannelId
	} else if r.Type != "multi" && bg.ReplyChannel.Id == "" {
		r.Channel = bg.ReplyChannel.Id
		r.Type = "command"
		checkMsg := strings.Split(r.Message, " ")
		if checkMsg[0] != "/echo" {
			r.Message = "/echo " + r.Message
		}
	}

	if "" == r.Channel {
		r.Channel = channelId
	}

	if r.Message != "" {
		switch r.Type {
		case "post":
			err = h.Mm.SendMsgToChannel(r.Message, r.Channel, post)
		case "command":
			err = h.Mm.SendCmdToChannel(r.Message, r.Channel, post)
		case "multi":
			messages := strings.Split(r.Message, "##")
			if len(messages) <= 1 {
				return fmt.Errorf("multi message wasn't formatted properly")
			}
			for _, m := range messages {
				messageParts := strings.Split(m, ";;")
				if len(messageParts) == 2 {
					u, _, _ := h.Mm.Client.GetUserByUsername(messageParts[0], "")
					c, _, _ := h.Mm.Client.CreateDirectChannel(u.Id, h.Mm.BotUser.Id)
					replyPost := &model.Post{}
					replyPost.ChannelId = c.Id
					replyPost.Message = messageParts[1]
					_, _, err = h.Mm.Client.CreatePost(replyPost)
				} else {
					err = h.Mm.SendMsgToChannel(m, r.Channel, post)
				}
			}

		case "dm":
			c, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = r.Message

			_, _, err = h.Mm.Client.CreatePost(replyPost)
		}
	}

	return err
}
