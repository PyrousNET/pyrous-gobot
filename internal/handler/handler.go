package handler

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
	"github.com/pyrousnet/pyrous-gobot/internal/users"

	"github.com/mattermost/mattermost-server/v6/model"
)

type Handler struct {
	Cache    cache.Cache
	Settings *settings.Settings
	Mm       *mmclient.MMClient
}

func NewHandler(mm *mmclient.MMClient, botCache cache.Cache) (*Handler, error) {
	settings, err := settings.NewSettings(mm.SettingsUrl)
	users.SetupUsers(mm, botCache)

	return &Handler{
		Settings: settings,
		Mm:       mm,
		Cache:    botCache,
	}, err
}

func (h *Handler) HandleWebSocketResponse(quit chan bool, event *model.WebSocketEvent) {
	if event.EventType() == "posted" {
		post := h.Mm.PostFromJson(strings.NewReader(event.GetData()["post"].(string)))

		// Ignore bot messages
		if post.UserId != h.Mm.BotUser.Id {
			if event.GetBroadcast().ChannelId == h.Mm.DebuggingChannel.Id {
				h.HandleMsgFromDebuggingChannel(event)
			} else {
				commandPattern := fmt.Sprintf(`^%s(.*)`, h.Settings.GetCommandTrigger())
				var triggerType string

				if ok, err := regexp.MatchString(commandPattern, post.Message); ok {
					if err != nil {
						log.Println(err)
						return
					}
					triggerType = "command"
				}

				gamePattern := `^\$(.*)`
				if ok, err := regexp.MatchString(gamePattern, post.Message); ok {
					if err != nil {
						log.Println(err)
						return
					}
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
	c, _, _ := h.Mm.Client.CreateDirectChannel(post.UserId, h.Mm.BotUser.Id)
	replyPost := &model.Post{}
	replyPost.ChannelId = c.Id
	replyPost.Message = message

	_, _, err := h.Mm.Client.CreatePost(replyPost)

	return err
}
