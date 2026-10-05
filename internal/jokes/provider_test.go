package jokes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubProvider struct {
	name  string
	joke  Joke
	err   error
	calls int
}

func (s *stubProvider) Name() string { return s.name }

func (s *stubProvider) Random(context.Context) (Joke, error) {
	s.calls++
	return s.joke, s.err
}

func TestCompositeUsesNextProviderAfterFailure(t *testing.T) {
	primary := &stubProvider{name: "reddit", err: errors.New("blocked")}
	backup := &stubProvider{name: "dad joke daily", joke: Joke{Setup: "Setup", Punchline: "Punchline"}}
	var failed []string

	got, err := (Composite{
		Providers: []Provider{primary, backup},
		OnFailure: func(name string, _ error) { failed = append(failed, name) },
	}).Random(context.Background())
	if err != nil {
		t.Fatalf("Random() error = %v", err)
	}
	if got != backup.joke {
		t.Fatalf("Random() = %#v, want %#v", got, backup.joke)
	}
	if primary.calls != 1 || backup.calls != 1 {
		t.Fatalf("provider calls = (%d, %d), want (1, 1)", primary.calls, backup.calls)
	}
	if len(failed) != 1 || failed[0] != "reddit" {
		t.Fatalf("failed providers = %v, want [reddit]", failed)
	}
}

func TestCompositeStopsAtFirstSuccessfulProvider(t *testing.T) {
	primary := &stubProvider{name: "reddit", joke: Joke{Setup: "Primary"}}
	backup := &stubProvider{name: "dad joke daily", joke: Joke{Setup: "Backup"}}

	got, err := (Composite{Providers: []Provider{primary, backup}}).Random(context.Background())
	if err != nil {
		t.Fatalf("Random() error = %v", err)
	}
	if got.Setup != "Primary" || backup.calls != 0 {
		t.Fatalf("Random() = %#v, backup calls = %d; want primary joke and no backup call", got, backup.calls)
	}
}

func TestCompositeReturnsErrorWhenAllProvidersFail(t *testing.T) {
	primary := &stubProvider{name: "reddit", err: errors.New("blocked")}
	backup := &stubProvider{name: "dad joke daily", err: errors.New("unavailable")}

	_, err := (Composite{Providers: []Provider{primary, backup}}).Random(context.Background())
	if err == nil {
		t.Fatal("Random() error = nil, want provider failures")
	}
	if got := err.Error(); got == "" || !containsAll(got, "reddit", "dad joke daily") {
		t.Fatalf("Random() error = %q, want both provider names", got)
	}
}

func TestRedditProviderSetsAuthHeadersAndParsesJoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Errorf("token Content-Type = %q", got)
			}
			username, password, ok := r.BasicAuth()
			if !ok || username != "client" || password != "secret" {
				t.Errorf("token BasicAuth = (%q, %q, %v)", username, password, ok)
			}
			if r.Header.Get("User-Agent") != "test-agent" {
				t.Errorf("token User-Agent = %q", r.Header.Get("User-Agent"))
			}
			_, _ = w.Write([]byte(`{"access_token":"test-token"}`))
			return
		}
		if r.URL.Path == "/feed" {
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Errorf("feed Authorization = %q", got)
			}
			_, _ = w.Write([]byte(`{"data":{"children":[{"data":{"title":"A setup","selftext":"A punchline"}}]}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	provider := Reddit{
		Client: server.Client(), TokenURL: server.URL + "/token", FeedURL: server.URL + "/feed",
		ClientID: "client", ClientSecret: "secret", UserAgent: "test-agent",
	}
	got, err := provider.Random(context.Background())
	if err != nil {
		t.Fatalf("Random() error = %v", err)
	}
	if got != (Joke{Setup: "A setup", Punchline: "A punchline"}) {
		t.Fatalf("Random() = %#v", got)
	}
}

func TestRedditProviderRejectsUnusableResponses(t *testing.T) {
	tests := []struct {
		name        string
		tokenStatus int
		tokenBody   string
		jokeStatus  int
		jokeBody    string
	}{
		{name: "HTML token response", tokenStatus: http.StatusOK, tokenBody: "<html>blocked</html>"},
		{name: "token missing access token", tokenStatus: http.StatusOK, tokenBody: `{"error":"unavailable"}`},
		{name: "token endpoint failure", tokenStatus: http.StatusForbidden, tokenBody: "network policy"},
		{name: "HTML joke response", tokenStatus: http.StatusOK, tokenBody: `{"access_token":"token"}`, jokeStatus: http.StatusOK, jokeBody: "<html>blocked</html>"},
		{name: "joke endpoint failure", tokenStatus: http.StatusOK, tokenBody: `{"access_token":"token"}`, jokeStatus: http.StatusServiceUnavailable, jokeBody: "unavailable"},
		{name: "empty feed", tokenStatus: http.StatusOK, tokenBody: `{"access_token":"token"}`, jokeStatus: http.StatusOK, jokeBody: `{"data":{"children":[]}}`},
		{name: "entry without content", tokenStatus: http.StatusOK, tokenBody: `{"access_token":"token"}`, jokeStatus: http.StatusOK, jokeBody: `{"data":{"children":[{"data":{}}]}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/token":
					w.WriteHeader(tt.tokenStatus)
					_, _ = w.Write([]byte(tt.tokenBody))
				case "/feed":
					w.WriteHeader(tt.jokeStatus)
					_, _ = w.Write([]byte(tt.jokeBody))
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			provider := Reddit{Client: server.Client(), TokenURL: server.URL + "/token", FeedURL: server.URL + "/feed"}
			if got, err := provider.Random(context.Background()); err == nil {
				t.Fatalf("Random() = %#v, error = nil; want failure", got)
			}
		})
	}
}

func TestDadJokeDailyProviderReturnsSplitJoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/jokes/random" || r.URL.Query().Get("type") != "dad" {
			t.Errorf("request URL = %s, want random dad joke endpoint", r.URL.String())
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("request User-Agent is empty")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"joke":{"setup":"Why did the scarecrow win an award?","punchline":"Because he was outstanding in his field!","type":"dad"}}`))
	}))
	defer server.Close()

	provider := DadJokeDaily{Client: server.Client(), Endpoint: server.URL + "/api/v1/jokes/random"}
	got, err := provider.Random(context.Background())
	if err != nil {
		t.Fatalf("Random() error = %v", err)
	}
	want := Joke{Setup: "Why did the scarecrow win an award?", Punchline: "Because he was outstanding in his field!"}
	if got != want {
		t.Fatalf("Random() = %#v, want %#v", got, want)
	}
}

func TestOpenAIProviderRequestsOriginalRobotStyleJoke(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/responses" {
			t.Errorf("request = %s %s, want POST /v1/responses", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization = %q, want bearer key", got)
		}
		var request struct {
			Model       string  `json:"model"`
			Input       string  `json:"input"`
			Temperature float64 `json:"temperature"`
			Text        struct {
				Format struct {
					Type   string `json:"type"`
					Name   string `json:"name"`
					Strict bool   `json:"strict"`
				} `json:"format"`
			} `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if request.Model != "test-model" || request.Temperature != 0.9 || request.Text.Format.Type != "json_schema" || !request.Text.Format.Strict {
			t.Errorf("request configuration = %#v, want test model, non-deterministic sampling, and strict JSON schema", request)
		}
		if !strings.Contains(strings.ToLower(request.Input), "bender") || !strings.Contains(strings.ToLower(request.Input), "original") {
			t.Errorf("prompt %q should request original Bender-inspired material", request.Input)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"output":[{"type":"message","content":[{"type":"output_text","text":"{\"setup\":\"Why did the robot open a comedy club?\",\"punchline\":\"The drinks were free, and it plans to steal the cash register.\"}"}]}]}`))
	}))
	defer server.Close()

	provider := OpenAI{Client: server.Client(), Endpoint: server.URL + "/v1/responses", APIKey: "test-key", Model: "test-model"}
	got, err := provider.Random(context.Background())
	if err != nil {
		t.Fatalf("Random() error = %v", err)
	}
	want := Joke{Setup: "Why did the robot open a comedy club?", Punchline: "The drinks were free, and it plans to steal the cash register."}
	if got != want {
		t.Fatalf("Random() = %#v, want %#v", got, want)
	}
}

func TestOpenAIProviderRejectsMalformedOrIncompleteResponse(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "missing output", body: `{"output":[]}`},
		{name: "invalid joke JSON", body: `{"output":[{"content":[{"type":"output_text","text":"not json"}]}]}`},
		{name: "missing punchline", body: `{"output":[{"content":[{"type":"output_text","text":"{\"setup\":\"Setup only\",\"punchline\":\" \"}"}]}]}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(tt.body)) }))
			defer server.Close()
			provider := OpenAI{Client: server.Client(), Endpoint: server.URL, APIKey: "test-key", Model: "test-model"}
			if joke, err := provider.Random(context.Background()); err == nil {
				t.Fatalf("Random() = %#v, error = nil; want failure", joke)
			}
		})
	}
}

func containsAll(value string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
