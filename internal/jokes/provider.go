package jokes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultUserAgent = "go:github.com/PyrousNET/pyrous-gobot:v1.0.0"

type Joke struct {
	Setup     string
	Punchline string
}

type Provider interface {
	Random(context.Context) (Joke, error)
}

type NamedProvider interface {
	Provider
	Name() string
}

// Composite tries providers in order and returns the first usable joke.
type Composite struct {
	Providers []Provider
	OnFailure func(provider string, err error)
}

func (providers Composite) Random(ctx context.Context) (Joke, error) {
	var failures []error
	for _, provider := range providers.Providers {
		if provider == nil {
			continue
		}
		joke, err := provider.Random(ctx)
		if err == nil && (strings.TrimSpace(joke.Setup) != "" || strings.TrimSpace(joke.Punchline) != "") {
			joke.Setup = strings.TrimSpace(joke.Setup)
			joke.Punchline = strings.TrimSpace(joke.Punchline)
			return joke, nil
		}
		if err == nil {
			err = errors.New("provider returned an empty joke")
		}
		name := "provider"
		if named, ok := provider.(NamedProvider); ok {
			name = named.Name()
		}
		if providers.OnFailure != nil {
			providers.OnFailure(name, err)
		}
		failures = append(failures, fmt.Errorf("%s: %w", name, err))
	}
	if len(failures) == 0 {
		return Joke{}, errors.New("no joke providers configured")
	}
	return Joke{}, errors.Join(failures...)
}

type Reddit struct {
	Client       *http.Client
	TokenURL     string
	FeedURL      string
	ClientID     string
	ClientSecret string
	UserAgent    string
}

type DadJokeDaily struct {
	Client    *http.Client
	Endpoint  string
	UserAgent string
}

// OpenAI generates an original two-part joke using the Responses API.
type OpenAI struct {
	Client   *http.Client
	Endpoint string
	APIKey   string
	Model    string
}

func (Reddit) Name() string       { return "reddit" }
func (DadJokeDaily) Name() string { return "dad joke daily" }
func (OpenAI) Name() string       { return "openai joke generator" }

const defaultOpenAIEndpoint = "https://api.openai.com/v1/responses"
const defaultOpenAIModel = "gpt-4.1-mini"

const jokeGenerationPrompt = `Write one original, family-friendly joke with a distinct setup and punchline. Its comedic voice should be inspired by broad traits associated with Bender from Futurama: a sarcastic, boastful, irreverent robot with selfish confidence. Do not quote or reuse dialogue, catchphrases, or jokes from the show. Keep each part short and suitable for a workplace chat. Return only the requested JSON.`

type openAIResponse struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (provider OpenAI) Random(ctx context.Context) (Joke, error) {
	if strings.TrimSpace(provider.APIKey) == "" {
		return Joke{}, errors.New("OpenAI API key is not configured")
	}
	endpoint := provider.Endpoint
	if endpoint == "" {
		endpoint = defaultOpenAIEndpoint
	}
	model := provider.Model
	if model == "" {
		model = defaultOpenAIModel
	}
	requestBody := struct {
		Model           string  `json:"model"`
		Input           string  `json:"input"`
		Temperature     float64 `json:"temperature"`
		MaxOutputTokens int     `json:"max_output_tokens"`
		Text            struct {
			Format struct {
				Type   string `json:"type"`
				Name   string `json:"name"`
				Strict bool   `json:"strict"`
				Schema struct {
					Type                 string                       `json:"type"`
					Properties           map[string]map[string]string `json:"properties"`
					Required             []string                     `json:"required"`
					AdditionalProperties bool                         `json:"additionalProperties"`
				} `json:"schema"`
			} `json:"format"`
		} `json:"text"`
	}{Model: model, Input: jokeGenerationPrompt, Temperature: 0.9, MaxOutputTokens: 180}
	requestBody.Text.Format.Type = "json_schema"
	requestBody.Text.Format.Name = "joke"
	requestBody.Text.Format.Strict = true
	requestBody.Text.Format.Schema.Type = "object"
	requestBody.Text.Format.Schema.Properties = map[string]map[string]string{
		"setup":     {"type": "string"},
		"punchline": {"type": "string"},
	}
	requestBody.Text.Format.Schema.Required = []string{"setup", "punchline"}
	requestBody.Text.Format.Schema.AdditionalProperties = false
	body, err := json.Marshal(requestBody)
	if err != nil {
		return Joke{}, fmt.Errorf("encode joke generation request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return Joke{}, fmt.Errorf("create joke generation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+provider.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	response, err := provider.client().Do(req)
	if err != nil {
		return Joke{}, fmt.Errorf("OpenAI joke generation request: %w", err)
	}
	defer response.Body.Close()
	if err := checkResponse("OpenAI joke generation request", response); err != nil {
		return Joke{}, err
	}
	var result openAIResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return Joke{}, fmt.Errorf("decode OpenAI joke response: %w", err)
	}
	for _, item := range result.Output {
		for _, content := range item.Content {
			if content.Type != "output_text" || strings.TrimSpace(content.Text) == "" {
				continue
			}
			var joke Joke
			if err := json.Unmarshal([]byte(content.Text), &joke); err != nil {
				return Joke{}, fmt.Errorf("decode generated joke: %w", err)
			}
			joke.Setup = strings.TrimSpace(joke.Setup)
			joke.Punchline = strings.TrimSpace(joke.Punchline)
			if joke.Setup == "" || joke.Punchline == "" {
				return Joke{}, errors.New("OpenAI returned an incomplete joke")
			}
			return joke, nil
		}
	}
	return Joke{}, errors.New("OpenAI response contained no joke text")
}

type redditAuth struct {
	AccessToken string `json:"access_token"`
}

type redditFeed struct {
	Data struct {
		Children []struct {
			Data struct {
				Title    string `json:"title"`
				Selftext string `json:"selftext"`
				Over18   bool   `json:"over_18"`
				Stickied bool   `json:"stickied"`
				IsVideo  bool   `json:"is_video"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (provider Reddit) Random(ctx context.Context) (Joke, error) {
	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, provider.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return Joke{}, fmt.Errorf("create reddit token request: %w", err)
	}
	req.SetBasicAuth(provider.ClientID, provider.ClientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", provider.userAgent())

	response, err := provider.client().Do(req)
	if err != nil {
		return Joke{}, fmt.Errorf("reddit token request: %w", err)
	}
	defer response.Body.Close()
	if err := checkResponse("reddit token request", response); err != nil {
		return Joke{}, err
	}

	var auth redditAuth
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&auth); err != nil {
		return Joke{}, fmt.Errorf("decode reddit token response: %w", err)
	}
	if auth.AccessToken == "" {
		return Joke{}, errors.New("reddit token response contained no access token")
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, provider.FeedURL, nil)
	if err != nil {
		return Joke{}, fmt.Errorf("create reddit feed request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+auth.AccessToken)
	req.Header.Set("User-Agent", provider.userAgent())
	response, err = provider.client().Do(req)
	if err != nil {
		return Joke{}, fmt.Errorf("reddit joke request: %w", err)
	}
	defer response.Body.Close()
	if err := checkResponse("reddit joke request", response); err != nil {
		return Joke{}, err
	}

	var feed redditFeed
	if err := json.NewDecoder(io.LimitReader(response.Body, 4<<20)).Decode(&feed); err != nil {
		return Joke{}, fmt.Errorf("decode reddit joke response: %w", err)
	}
	if len(feed.Data.Children) == 0 {
		return Joke{}, errors.New("reddit returned no jokes")
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(feed.Data.Children), func(i, j int) {
		feed.Data.Children[i], feed.Data.Children[j] = feed.Data.Children[j], feed.Data.Children[i]
	})
	for _, child := range feed.Data.Children {
		item := child.Data
		if item.Over18 || item.Stickied || item.IsVideo {
			continue
		}
		setup, punchline := strings.TrimSpace(item.Title), strings.TrimSpace(item.Selftext)
		if setup == "" {
			setup, punchline = punchline, ""
		}
		if setup != "" || punchline != "" {
			return Joke{Setup: setup, Punchline: punchline}, nil
		}
	}
	return Joke{}, errors.New("reddit returned no usable jokes")
}

func (provider DadJokeDaily) Random(ctx context.Context) (Joke, error) {
	endpoint := provider.Endpoint
	if endpoint == "" {
		endpoint = "https://dadjoke.win/api/v1/jokes/random"
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return Joke{}, fmt.Errorf("parse dad joke daily endpoint: %w", err)
	}
	query := parsed.Query()
	query.Set("type", "dad")
	parsed.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return Joke{}, fmt.Errorf("create dad joke daily request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", provider.userAgent())
	response, err := provider.client().Do(req)
	if err != nil {
		return Joke{}, fmt.Errorf("dad joke daily request: %w", err)
	}
	defer response.Body.Close()
	if err := checkResponse("dad joke daily request", response); err != nil {
		return Joke{}, err
	}
	var payload struct {
		Joke Joke `json:"joke"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&payload); err != nil {
		return Joke{}, fmt.Errorf("decode dad joke daily response: %w", err)
	}
	if strings.TrimSpace(payload.Joke.Setup) == "" && strings.TrimSpace(payload.Joke.Punchline) == "" {
		return Joke{}, errors.New("dad joke daily returned an empty joke")
	}
	return Joke{Setup: strings.TrimSpace(payload.Joke.Setup), Punchline: strings.TrimSpace(payload.Joke.Punchline)}, nil
}

func (provider Reddit) client() *http.Client {
	if provider.Client != nil {
		return provider.Client
	}
	return http.DefaultClient
}

func (provider DadJokeDaily) client() *http.Client {
	if provider.Client != nil {
		return provider.Client
	}
	return http.DefaultClient
}

func (provider OpenAI) client() *http.Client {
	if provider.Client != nil {
		return provider.Client
	}
	return http.DefaultClient
}

func (provider Reddit) userAgent() string {
	if provider.UserAgent != "" {
		return provider.UserAgent
	}
	return defaultUserAgent
}

func (provider DadJokeDaily) userAgent() string {
	if provider.UserAgent != "" {
		return provider.UserAgent
	}
	return defaultUserAgent
}

func checkResponse(operation string, response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	const maxBodyLength = 512
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBodyLength+1))
	detail := strings.TrimSpace(string(body))
	if len(detail) > maxBodyLength {
		detail = detail[:maxBodyLength] + "..."
	}
	if err != nil {
		return fmt.Errorf("%s returned HTTP %s (could not read response body: %v)", operation, response.Status, err)
	}
	if detail == "" {
		return fmt.Errorf("%s returned HTTP %s", operation, response.Status)
	}
	return fmt.Errorf("%s returned HTTP %s: %q", operation, response.Status, detail)
}
