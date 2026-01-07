package handler

import (
	"context"
	"encoding/json"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/pubsub"
	"log"
	"regexp"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"github.com/pyrousnet/pyrous-gobot/internal/users"

	"github.com/mattermost/mattermost/server/public/model"
)

type Handler struct {
	Cache           cache.Cache
	Pubsub          pubsub.Pubsub
	Settings        *settings.Settings
	Mm              *mmclient.MMClient
	ResponseChannel chan comms.Response
}

func NewHandler(mm *mmclient.MMClient, botCache cache.Cache, botPubSub pubsub.Pubsub) (*Handler, error) {
	settings, err := settings.NewSettings(mm.SettingsUrl)
	users.SetupUsers(mm, botCache)

	rb, ok, err := botCache.Get("sys_restarted_by_user")
	if ok {
		var rm map[string]string
		json.Unmarshal([]byte(rb.(string)), &rm)

		replyPost := &model.Post{}

		mm.SendMsgToChannel("I'm back, baby!", rm["channel"], replyPost)

		c, _, err := mm.Client.CreateDirectChannel(context.Background(), rm["user"], mm.BotUser.Id)
		if err != nil {
			log.Print(err)
		}

		replyPost.ChannelId = c.Id
		replyPost.Message = "See?  😉"

		_, _, err = mm.Client.CreatePost(context.Background(), replyPost)
		if err != nil {
			log.Print(err)
		}

		botCache.Clean("sys_restarted_by_user")
	}
	mRH := comms.MessageHandler{
		Mm:         mm,
		ResponseCh: make(chan comms.Response, 50),
		Cache:      botCache,
	}

	h := &Handler{
		Settings:        settings,
		Mm:              mm,
		Cache:           botCache,
		Pubsub:          botPubSub,
		ResponseChannel: mRH.ResponseCh,
	}

	mRH.StartMessageHandler()

	return h, err
}

func (h *Handler) HandleWebSocketResponse(quit chan bool, event *model.WebSocketEvent) {
	if event.EventType() == "posted" {
		post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))

		// Ignore bot messages
		if post.UserId != h.Mm.BotUser.Id {
			if event.GetBroadcast().ChannelId == h.Mm.DebuggingChannel.Id {
				h.HandleMsgFromDebuggingChannel(event)
			} else {
				var triggerType string
				if strings.HasPrefix(post.Message, "!") {
					triggerType = "command"
				} else if strings.HasPrefix(post.Message, "$") {
					triggerType = "game"
				}

				h.HandleMsgFromChannel(triggerType, quit, event)
			}
		}
	}
}

func (h *Handler) HandleMsgFromChannel(triggerType string, quit chan bool, event *model.WebSocketEvent) {
	switch triggerType {
	case "command":
		err := h.HandleCommand(quit, event)
		if err != nil {
			log.Println(err)
		}
	case "game":
		err := h.HandleGame(quit, event)
		if err != nil {
			log.Println(err)
		}
	default:
		post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))

		err := users.HandlePost(post, h.Mm, h.Cache)
		if err != nil {
			log.Println(err)
		}
	}
}

func (h *Handler) HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// Lets only reponded to messaged posted events
	println("responding to debugging channel msg")

	post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	if post != nil {
		// if you see any word matching 'alive' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
			h.Mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'up' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)up(?:$|\W)`, post.Message); matched {
			h.Mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'running' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)running(?:$|\W)`, post.Message); matched {
			h.Mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'hello' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {
			h.Mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}
	}

	h.Mm.SendMsgToDebuggingChannel("I did not understand you!", post.Id)
}

func (h *Handler) SendErrorResponse(post *model.Post, message string) error {
	c, _, _ := h.Mm.Client.CreateDirectChannel(context.Background(), post.UserId, h.Mm.BotUser.Id)
	replyPost := &model.Post{}
	replyPost.ChannelId = c.Id
	replyPost.Message = message

	_, _, err := h.Mm.Client.CreatePost(context.Background(), replyPost)

	return err
}
