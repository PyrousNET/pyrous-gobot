package handler

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/commands"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"

	"github.com/mattermost/mattermost-server/v5/model"
)

type Handler struct {
	Cache    cache.Cache
	Settings *settings.Settings
	mm       *mmclient.MMClient
}

func NewHandler(mm *mmclient.MMClient, redis cache.Cache) (*Handler, error) {
	settings, err := settings.NewSettings(mm.SettingsUrl)

	return &Handler{
		Settings: settings,
		mm:       mm,
		Cache:    redis,
	}, err
}

func (h *Handler) HandleWebSocketResponse(quit chan bool, event *model.WebSocketEvent) {
	if event.GetBroadcast().ChannelId == h.mm.DebuggingChannel.Id {
		h.HandleMsgFromDebuggingChannel(event)
	} else {
		h.HandleMsgFromChannel(quit, event)
	}
}

func (h *Handler) HandleMsgFromChannel(quit chan bool, event *model.WebSocketEvent) {
	//Only handle messaged posted events
	if event.EventType() != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	cmds := commands.NewCommands(h.Settings, h.mm)

	channelId := event.GetBroadcast().ChannelId
	post := model.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))

	// Ignore bot messages
	if post.UserId == h.mm.BotUser.Id {
		return
	}

	pattern := fmt.Sprintf(`^%s(.*)`, h.Settings.GetCommandTrigger())

	ok, err := regexp.MatchString(pattern, post.Message)
	if ok {
		response, err := cmds.HandleCommandMsgFromWebSocket(event)
		if err == nil {
			if "" == response.Channel {
				response.Channel = channelId
			}

			if response.Type != "shutdown" {
				dmchannel, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
				if response.Channel == dmchannel.Id {
					response.Type = "dm"
				}
			}

			if response.Message != "" {
				switch response.Type {
				case "post":
					err = h.mm.SendMsgToChannel(response.Message, response.Channel, post)
				case "command":
					err = h.mm.SendCmdToChannel(response.Message, response.Channel, post)
				case "dm":
					c, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
					post := &model.Post{}
					post.ChannelId = c.Id
					post.Message = response.Message

					_, e := h.mm.Client.CreatePost(post)
					if e.Error != nil {
						err = fmt.Errorf("%+v\n", e.Error)
					}
				case "shutdown":
					c, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
					post := &model.Post{}
					post.ChannelId = c.Id
					post.Message = response.Message

					_, e := h.mm.Client.CreatePost(post)
					if e.Error != nil {
						err = fmt.Errorf("%+v\n", e.Error)
					}

					quit <- true
				}
			}
		} else {
			log.Println(err)
		}
	}

	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// Lets only reponded to messaged posted events
	if event.EventType() != model.WEBSOCKET_EVENT_POSTED {
		return
	}

	println("responding to debugging channel msg")

	post := model.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
	if post != nil {

		// ignore my events
		if post.UserId == h.mm.BotUser.Id {
			return
		}

		// if you see any word matching 'alive' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)alive(?:$|\W)`, post.Message); matched {
			h.mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'up' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)up(?:$|\W)`, post.Message); matched {
			h.mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'running' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)running(?:$|\W)`, post.Message); matched {
			h.mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}

		// if you see any word matching 'hello' then respond
		if matched, _ := regexp.MatchString(`(?:^|\W)hello(?:$|\W)`, post.Message); matched {
			h.mm.SendMsgToDebuggingChannel("Yes I'm running", post.Id)
			return
		}
	}

	h.mm.SendMsgToDebuggingChannel("I did not understand you!", post.Id)
}
