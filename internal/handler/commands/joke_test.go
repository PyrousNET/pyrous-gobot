package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/jokes"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
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

func TestFetchJokeOffersFeedbackForAnyProvider(t *testing.T) {
	c := cache.GetLocalCache()
	responses := make(chan comms.Response, 2)
	event := BotCommand{cache: c, ResponseChannel: responses}
	provider := &testJokeProvider{joke: jokes.Joke{Setup: "A setup", Punchline: "A punchline", Provider: "reddit"}}
	response := comms.Response{ReplyChannelId: "channel-id", UserId: "user-id"}

	if err := fetchJoke(event, response, provider); err != nil {
		t.Fatalf("fetchJoke() error = %v", err)
	}
	setup, delivery := <-responses, <-responses
	if strings.Contains(setup.Message, "!jokefeedback") {
		t.Fatalf("setup should not ask for feedback before the punchline: %q", setup.Message)
	}
	punchlineIndex := strings.Index(delivery.Message, "A punchline")
	feedbackIndex := strings.Index(delivery.Message, "!jokefeedback <rating>")
	if punchlineIndex < 0 || feedbackIndex <= punchlineIndex {
		t.Fatalf("feedback instructions should follow the punchline: %q", delivery.Message)
	}
	value, ok, err := c.Get(pendingJokeFeedbackKey("channel-id", "user-id"))
	if err != nil || !ok {
		t.Fatalf("pending joke lookup = (%v, %v), want a saved joke; error = %v", value, ok, err)
	}
	var pending pendingJokeFeedback
	if err := json.Unmarshal([]byte(value.(string)), &pending); err != nil {
		t.Fatal(err)
	}
	if pending.Joke.Setup != "A setup" || pending.Joke.Punchline != "A punchline" || pending.Joke.Provider != "reddit" {
		t.Fatalf("pending joke = %#v", pending.Joke)
	}
}

func TestJokefeedbackAppendsRatingAndCommentToCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "feedback.csv")
	t.Setenv("JOKE_FEEDBACK_CSV", path)
	c := cache.GetLocalCache()
	userData, err := json.Marshal(users.User{Id: "user-id", Name: "alice"})
	if err != nil {
		t.Fatal(err)
	}
	c.Put(users.KeyPrefix+"alice", string(userData))
	pending, err := json.Marshal(pendingJokeFeedback{Joke: jokes.Joke{Setup: "A setup", Punchline: "A punchline", Provider: jokes.OpenAIProviderName}})
	if err != nil {
		t.Fatal(err)
	}
	c.Put(pendingJokeFeedbackKey("channel-id", "user-id"), string(pending))
	responses := make(chan comms.Response, 1)
	event := BotCommand{
		body:            "2 The punchline did not follow",
		sender:          "@alice",
		ReplyChannel:    &model.Channel{Id: "channel-id"},
		ResponseChannel: responses,
		cache:           c,
	}
	if err := (BotCommand{}).Jokefeedback(event); err != nil {
		t.Fatalf("Jokefeedback() error = %v", err)
	}
	if response := <-responses; !strings.Contains(response.Message, "2/5") {
		t.Fatalf("feedback response = %#v", response)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	rows, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][1] != "2" || rows[1][2] != "The punchline did not follow" || rows[1][3] != "A setup" || rows[1][4] != "A punchline" {
		t.Fatalf("feedback CSV rows = %#v", rows)
	}
	if _, ok, _ := c.Get(pendingJokeFeedbackKey("channel-id", "user-id")); ok {
		t.Fatal("pending joke should be cleared after a rating is saved")
	}
}

func TestJokefeedbackRejectsRatingsOutsideOneToFive(t *testing.T) {
	responses := make(chan comms.Response, 1)
	event := BotCommand{body: "6", ReplyChannel: &model.Channel{Id: "channel-id"}, ResponseChannel: responses}
	if err := (BotCommand{}).Jokefeedback(event); err != nil {
		t.Fatalf("Jokefeedback() error = %v", err)
	}
	if response := <-responses; !strings.Contains(response.Message, "from 1 to 5") {
		t.Fatalf("invalid-rating response = %#v", response)
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
