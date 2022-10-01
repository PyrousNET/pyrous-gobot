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
		case "multi": // TODO - We need to rethink this. It only allows 2 commands.
			messages := strings.Split(r.Message, "##")
			var firstMessage, secondMessage string
			if len(messages) > 1 {
				firstMessage = messages[0]
				secondMessage = messages[1]
			} else {
				return fmt.Errorf("multi message wasn't formatted properly")
			}
			if err != nil {
				return err
			}
			c, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = firstMessage

			_, _, err = h.Mm.Client.CreatePost(replyPost)
			err = h.Mm.SendMsgToChannel(secondMessage, r.Channel, post)
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
