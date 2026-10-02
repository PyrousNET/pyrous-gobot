package commands

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
)

func TestFetchJokeHandlesUnusableRedditResponses(t *testing.T) {
	tests := []struct {
		name        string
		tokenStatus int
		tokenBody   string
		jokeStatus  int
		jokeBody    string
	}{
		{
			name:        "HTML token response",
			tokenStatus: http.StatusOK,
			tokenBody:   "<html>temporarily unavailable</html>",
		},
		{
			name:        "token response without access token",
			tokenStatus: http.StatusOK,
			tokenBody:   `{"error":"temporarily unavailable"}`,
		},
		{
			name:        "token endpoint failure",
			tokenStatus: http.StatusServiceUnavailable,
			tokenBody:   "upstream unavailable",
		},
		{
			name:        "HTML joke response",
			tokenStatus: http.StatusOK,
			tokenBody:   `{"access_token":"test-token"}`,
			jokeStatus:  http.StatusOK,
			jokeBody:    "<html>rate limited</html>",
		},
		{
			name:        "joke endpoint failure",
			tokenStatus: http.StatusOK,
			tokenBody:   `{"access_token":"test-token"}`,
			jokeStatus:  http.StatusServiceUnavailable,
			jokeBody:    "upstream unavailable",
		},
		{
			name:        "empty joke feed",
			tokenStatus: http.StatusOK,
			tokenBody:   `{"access_token":"test-token"}`,
			jokeStatus:  http.StatusOK,
			jokeBody:    `{"data":{"children":[]}}`,
		},
		{
			name:        "joke entry without content",
			tokenStatus: http.StatusOK,
			tokenBody:   `{"access_token":"test-token"}`,
			jokeStatus:  http.StatusOK,
			jokeBody:    `{"data":{"children":[{"data":{}}]}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/token":
					w.WriteHeader(tt.tokenStatus)
					_, _ = w.Write([]byte(tt.tokenBody))
				case "/jokes":
					w.WriteHeader(tt.jokeStatus)
					_, _ = w.Write([]byte(tt.jokeBody))
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			responses := make(chan comms.Response, 2)
			event := BotCommand{
				ReplyChannel:    &model.Channel{Id: "test-channel"},
				ResponseChannel: responses,
			}
			err := fetchJoke(event, comms.Response{ReplyChannelId: "test-channel"}, server.Client(), server.URL+"/token", server.URL+"/jokes")
			if err != nil {
				t.Fatalf("fetchJoke() error = %v, want nil after sending a failure response", err)
			}

			select {
			case response := <-responses:
				if response.Type != "dm" {
					t.Fatalf("response type = %q, want dm", response.Type)
				}
				if !strings.Contains(response.Message, "couldn't fetch a joke") {
					t.Fatalf("response message = %q, want friendly fetch failure", response.Message)
				}
			default:
				t.Fatal("fetchJoke() did not send a failure response")
			}
			if len(responses) != 0 {
				t.Fatalf("fetchJoke() sent %d unexpected extra responses", len(responses))
			}
		})
	}
}

func TestFetchJokePostsParsedJoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			_, _ = w.Write([]byte(`{"access_token":"test-token"}`))
			return
		}
		if r.URL.Path == "/jokes" {
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Errorf("Authorization = %q, want Bearer test-token", got)
			}
			_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"title":"A dad joke","selftext":"A punchline"}}]}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	responses := make(chan comms.Response, 2)
	event := BotCommand{ResponseChannel: responses}
	err := fetchJoke(event, comms.Response{}, server.Client(), server.URL+"/token", server.URL+"/jokes")
	if err != nil {
		t.Fatalf("fetchJoke() error = %v", err)
	}

	title := <-responses
	if title.Type != "post" || title.Message != "A dad joke" {
		t.Fatalf("title response = %#v, want joke title post", title)
	}
	punchline := <-responses
	if punchline.Type != "command" || punchline.Message != `/echo "A punchline" 5` {
		t.Fatalf("punchline response = %#v, want echo command", punchline)
	}
}
