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

func TestHandleCommand_ChannelResolution(t *testing.T) {
	tests := []struct {
		name               string
		broadcastChannelID string
		postChannelID      string
	}{
		{
			name:               "uses broadcast channel when present",
			broadcastChannelID: "chan123",
			postChannelID:      "postchan",
		},
		{
			name:               "uses post channel when broadcast missing",
			broadcastChannelID: "",
			postChannelID:      "chan123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channelID := tt.postChannelID
			if tt.broadcastChannelID != "" {
				channelID = tt.broadcastChannelID
			}

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
				}),
				Mm: &mmclient.MMClient{
					Client: client,
				},
				Cache:           &cache.MockCache{},
				ResponseChannel: make(chan comms.Response, 1),
			}

			event := model.NewWebSocketEvent(model.WebsocketEventPosted, "", tt.broadcastChannelID, "", nil, "")
			post := &model.Post{UserId: "user1", ChannelId: tt.postChannelID, Message: "!help"}
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
			case <-time.After(500 * time.Millisecond):
				t.Fatal("expected help response on channel")
			}
		})
	}
}
