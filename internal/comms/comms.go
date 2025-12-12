package comms

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
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
	ctx        context.Context
}

func (h *MessageHandler) StartMessageHandler() {
	h.ctx = context.Background()
	go func() {
		for {
			r := <-h.ResponseCh
			go h.SendMessage(&r)
		}
	}()

}

func (h *MessageHandler) SendMessage(r *Response) {
	start := time.Now()
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

	dmchannel, _, dmErr := h.Mm.Client.CreateDirectChannel(h.ctx, h.Mm.BotUser.Id, r.UserId)
	if dmErr != nil {
		log.Error(dmErr)
	}
	if r.Type != "shutdown" {
		if dmchannel != nil && r.ReplyChannelId == dmchannel.Id {
			r.Type = "dm"
		}
	}

	if r.Type == "dm" {
		if strings.HasPrefix(checkMsg[0], "/") {
			commandParts := strings.Split(r.Message, "\"")
			if len(commandParts) > 1 && commandParts[1] != "" {
				post.Message = commandParts[1]
			}
		}

	}

	if r.Message != "" {
		err = h.sendWithRetry(r, post, dmchannel)
	}

	stdlog.Printf("[msg] type=%s channel=%s user=%s len=%d dur=%s err=%v", r.Type, r.ReplyChannelId, r.UserId, len(r.Message), time.Since(start), err)
}

func (h *MessageHandler) sendWithRetry(r *Response, post *model.Post, dmchannel *model.Channel) error {
	const maxAttempts = 3
	backoff := 200 * time.Millisecond
	var err error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		switch r.Type {
		case "post":
			err = h.Mm.SendMsgToChannel(r.Message, r.ReplyChannelId, post)
		case "command":
			err = h.Mm.SendCmdToChannel(r.Message, r.ReplyChannelId, post)
		case "dm":
			if dmchannel == nil {
				err = fmt.Errorf("unable to open DM channel for user %s", r.UserId)
				log.Error(err)
				return err
			}

			post.ChannelId = dmchannel.Id
			if post.Message[0] == '/' {
				_, _, err = h.Mm.Client.ExecuteCommandWithTeam(h.ctx, post.ChannelId, h.Mm.BotTeam.Id, post.Message)
			} else {
				_, _, err = h.Mm.Client.CreatePost(h.ctx, post)
			}
		case "shutdown":
			c, _, errShutdown := h.Mm.Client.CreateDirectChannel(h.ctx, r.UserId, h.Mm.BotUser.Id)
			if errShutdown != nil {
				log.Error(errShutdown)
			}
			replyPost := &model.Post{}
			replyPost.ChannelId = c.Id
			replyPost.Message = r.Message

			if _, _, errShutdown = h.Mm.Client.CreatePost(h.ctx, replyPost); errShutdown != nil {
				log.Error(errShutdown)
			}

			if errShutdown = h.Mm.SendMsgToChannel("Awe, Crap!", r.ReplyChannelId, post); errShutdown != nil {
				log.Error(errShutdown)
			}

			cache := map[string]interface{}{
				"user":    post.UserId,
				"channel": r.ReplyChannelId,
			}
			cj, _ := json.Marshal(cache)

			h.Cache.Put("sys_restarted_by_user", cj)

			r.Quit <- true
			return errShutdown
		default:
			err = fmt.Errorf("unknown message type %s", r.Type)
		}

		if err == nil {
			return nil
		}

		if attempt < maxAttempts {
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	log.Error(err)
	return err
}
