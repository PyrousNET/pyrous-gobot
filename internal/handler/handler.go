package handler

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/handler/commands"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"github.com/pyrousnet/pyrous-gobot/internal/users"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Handler struct {
	Cache    cache.Cache
	Settings *settings.Settings
	mm       *mmclient.MMClient
}

func NewHandler(mm *mmclient.MMClient, botCache cache.Cache) (*Handler, error) {
	settings, err := settings.NewSettings(mm.SettingsUrl)
	users.SetupUsers(mm, botCache)

	return &Handler{
		Settings: settings,
		mm:       mm,
		Cache:    botCache,
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
	if event.EventType() != "posted" {
		return
	}

	cmds := commands.NewCommands(h.Settings, h.mm, h.Cache)

	channelId := event.GetBroadcast().ChannelId
	post := h.mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))

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
				dmchannel, _, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
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
					c, _, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
					replyPost := &model.Post{}
					replyPost.ChannelId = c.Id
					replyPost.Message = response.Message

					_, _, err = h.mm.Client.CreatePost(replyPost)

				case "shutdown":
					c, _, _ := h.mm.Client.CreateDirectChannel(post.UserId, h.mm.BotUser.Id)
					replyPost := &model.Post{}
					replyPost.ChannelId = c.Id
					replyPost.Message = response.Message

					_, _, err := h.mm.Client.CreatePost(replyPost)

					err = h.mm.SendMsgToChannel("Awe, Crap!", response.Channel, post)
					if err != nil {
						log.Print(err)
					}

					quit <- true
				}
			}
		} else {
			log.Println(err)
		}
	} else {
		users.HandlePost(post, h.mm, h.Cache)
	}

	if err != nil {
		log.Println(err)
	}
}

func (h *Handler) HandleMsgFromDebuggingChannel(event *model.WebSocketEvent) {
	// Lets only reponded to messaged posted events
	if event.EventType() != "posted" {
		return
	}

	println("responding to debugging channel msg")

	post := h.mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))
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
