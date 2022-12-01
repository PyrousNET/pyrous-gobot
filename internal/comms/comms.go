package comms

import (
	"encoding/json"
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"strings"
)

type Response struct {
	ReplyChannelId string
	Message        string
	Type           string
	UserId         string
	Quit           chan bool
}

type MessageHandler struct {
	ResponseCh chan Response
	Mm         *mmclient.MMClient
	Cache      cache.Cache
}

func (h *MessageHandler) StartMessageHandler() {
	go func() {
		for {
			r := <-h.ResponseCh
			go h.SendMessage(&r)
		}
	}()

}

func (h *MessageHandler) SendMessage(r *Response) {
	post := &model.Post{
		ChannelId: r.ReplyChannelId,
		Message:   r.Message,
		UserId:    r.UserId,
	}
	var err error
	checkMsg := strings.Split(r.Message, " ")
	if r.Type == "command" {
		if checkMsg[0] != "/echo" && !strings.HasPrefix(checkMsg[0], "/") {
			r.Message = "/echo " + r.Message
		}
	}

	if r.Type != "shutdown" {
		dmchannel, _, _ := h.Mm.Client.CreateDirectChannel(r.UserId, h.Mm.BotUser.Id)
		if r.ReplyChannelId == dmchannel.Id {
			r.Type = "dm"
		}
	}

	if r.Message != "" {
		switch r.Type {
		case "post":
			err = h.Mm.SendMsgToChannel(r.Message, r.ReplyChannelId, post)
		case "command":
			err = h.Mm.SendCmdToChannel(r.Message, r.ReplyChannelId, post)
		case "dm":
			c, _, err := h.Mm.Client.CreateDirectChannel(r.UserId, h.Mm.BotUser.Id)
			if err != nil {
				panic(err)
			}

			post.ChannelId = c.Id
			_, _, err = h.Mm.Client.CreatePost(post)
		case "shutdown":
			c, _, err := h.Mm.Client.CreateDirectChannel(r.UserId, h.Mm.BotUser.Id)
			if err != nil {
				panic(err)
			}
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = r.Message

			_, _, err = h.Mm.Client.CreatePost(replyPost)

			err = h.Mm.SendMsgToChannel("Awe, Crap!", r.ReplyChannelId, post)
			if err != nil {
				log.Error(err)
			}

			cache := map[string]interface{}{
				"user":    post.UserId,
				"channel": r.ReplyChannelId,
			}
			cj, _ := json.Marshal(cache)

			h.Cache.Put("sys_restarted_by_user", cj)

			r.Quit <- true
		}

		if err != nil {
			log.Error(err)
		}
	}
}
