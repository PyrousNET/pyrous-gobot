package comms

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
)

type Response struct {
	ReplyChannelId string
	Message        string
	Type           string
	UserId         string
}

type MessageHandler struct {
	ResponseCh chan Response
	Mm         *mmclient.MMClient
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
	post := &model.Post{}
	post.ChannelId = r.ReplyChannelId
	post.Message = r.Message
	var err error
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
		}

		if err != nil {
			log.Error(err)
		}
	}
}
