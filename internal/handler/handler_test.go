package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestHandleWebSocketResponse_IgnoresNonCommandMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v4/users/") {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(&model.User{Id: "user1", Username: "user1"})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := model.NewAPIv4Client(server.URL)
	h := &Handler{
		Settings: settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
			CommandTrigger: "!",
			GameTrigger:    "$",
		}),
		Mm: &mmclient.MMClient{
			Client:           client,
			BotUser:          &model.User{Id: "bot"},
			DebuggingChannel: &model.Channel{Id: "debug"},
		},
		Cache: &cache.MockCache{},
	}

	event := model.NewWebSocketEvent(model.WebsocketEventPosted, "", "chan", "", nil, "")
	post := &model.Post{UserId: "user1", Message: "just chatting"}
	b, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal post: %v", err)
	}
	event.Add("post", string(b))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	h.HandleWebSocketResponse(make(chan bool), event)
}

func TestHandleCommand_HelpCommandResponds(t *testing.T) {
	channelID := "chan123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v4/channels/"+channelID) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(&model.Channel{Id: channelID})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := model.NewAPIv4Client(server.URL)
	h := &Handler{
		Settings: settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{
			CommandTrigger: "!",
			GameTrigger:    "$",
		}),
		Mm: &mmclient.MMClient{
			Client: client,
		},
		Cache:           &cache.MockCache{},
		ResponseChannel: make(chan comms.Response, 1),
	}

	event := model.NewWebSocketEvent(model.WebsocketEventPosted, "", channelID, "", nil, "")
	post := &model.Post{UserId: "user1", Message: "!help"}
	b, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal post: %v", err)
	}
	event.Add("post", string(b))
	event.Add("sender_name", "@tester")

	if err := h.HandleCommand(make(chan bool), event); err != nil {
		t.Fatalf("HandleCommand error: %v", err)
	}

	select {
	case resp := <-h.ResponseChannel:
		if resp.ReplyChannelId != channelID {
			t.Fatalf("unexpected reply channel: %q", resp.ReplyChannelId)
		}
		if !strings.Contains(resp.Message, "Available commands") {
			t.Fatalf("unexpected help response: %q", resp.Message)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected help response on channel")
	}
}
