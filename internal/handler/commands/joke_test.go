package commands

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/jokes"
)

type testJokeProvider struct {
	joke  jokes.Joke
	err   error
	calls int
}

func (provider *testJokeProvider) Name() string { return "test" }

func (provider *testJokeProvider) Random(context.Context) (jokes.Joke, error) {
	provider.calls++
	return provider.joke, provider.err
}

func TestFetchJokePostsSetupThenDelayedPunchline(t *testing.T) {
	responses := make(chan comms.Response, 2)
	event := BotCommand{ResponseChannel: responses}
	provider := &testJokeProvider{joke: jokes.Joke{Setup: "A setup", Punchline: "A punchline"}}

	if err := fetchJoke(event, comms.Response{}, provider); err != nil {
		t.Fatalf("fetchJoke() error = %v", err)
	}
	setup, delivery := <-responses, <-responses
	if setup.Type != "post" || setup.Message != "A setup" {
		t.Fatalf("setup response = %#v, want setup post", setup)
	}
	if delivery.Type != "command" || delivery.Message != `/echo "A punchline" 5` {
		t.Fatalf("delivery response = %#v, want delayed punchline", delivery)
	}
}

func TestRedditFailureUsesDadJokeDailyBackup(t *testing.T) {
	dadJokeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/jokes/random" || r.URL.Query().Get("type") != "dad" {
			t.Errorf("backup request URL = %s, want dad joke random endpoint", r.URL.String())
		}
		_, _ = w.Write([]byte(`{"joke":{"setup":"Why did the scarecrow win an award?","punchline":"Because he was outstanding in his field!","type":"dad"}}`))
	}))
	defer dadJokeServer.Close()

	redditServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			t.Errorf("Reddit request path = %q, want /token", r.URL.Path)
		}
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("blocked by network policy"))
	}))
	defer redditServer.Close()

	responses := make(chan comms.Response, 3)
	event := BotCommand{ResponseChannel: responses}
	providers := jokes.Composite{Providers: []jokes.Provider{
		jokes.Reddit{Client: redditServer.Client(), TokenURL: redditServer.URL + "/token", FeedURL: redditServer.URL + "/feed", ClientID: "client", ClientSecret: "secret"},
		jokes.DadJokeDaily{Client: dadJokeServer.Client(), Endpoint: dadJokeServer.URL + "/api/v1/jokes/random"},
	}}
	if err := fetchJoke(event, comms.Response{}, providers); err != nil {
		t.Fatalf("fetchJoke() error = %v", err)
	}
	setup, delivery := <-responses, <-responses
	if setup.Type != "post" || setup.Message != "Why did the scarecrow win an award?" {
		t.Fatalf("setup response = %#v, want backup setup post", setup)
	}
	if delivery.Type != "command" || delivery.Message != `/echo "Because he was outstanding in his field!" 5` {
		t.Fatalf("delivery response = %#v, want delayed backup punchline", delivery)
	}
	if len(responses) != 0 {
		t.Fatalf("got %d unexpected responses; successful backup should not send failure DM", len(responses))
	}
}

func TestFetchJokeSendsDMAndSillyPostWhenAllProvidersFail(t *testing.T) {
	responses := make(chan comms.Response, 2)
	event := BotCommand{
		ReplyChannel:    &model.Channel{Id: "test-channel"},
		ResponseChannel: responses,
	}
	provider := &testJokeProvider{err: errors.New("upstream unavailable")}
	if err := fetchJoke(event, comms.Response{ReplyChannelId: "test-channel"}, provider); err != nil {
		t.Fatalf("fetchJoke() error = %v", err)
	}
	dm, post := <-responses, <-responses
	if dm.Type != "dm" || dm.ReplyChannelId != "test-channel" {
		t.Fatalf("failure DM = %#v", dm)
	}
	if post.Type != "post" || post.Message != dadJokeFallback {
		t.Fatalf("fallback post = %#v, want local silly joke", post)
	}
}
