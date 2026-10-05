package commands

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/jokes"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const (
	redditTokenURL    = "https://www.reddit.com/api/v1/access_token"
	redditDadJokesURL = "https://oauth.reddit.com/r/dadjokes"
	redditClientID    = "aIuZxRUiUiPIFD-fVb--jg"
	redditSecret      = "UpGXB262RUsADk1RNU3vaMqLFCKxmQ"
	redditUserAgent   = "go:github.com/PyrousNET/pyrous-gobot:v1.0.0"
	dadJokeFallback   = "The joke suppliers are taking a nap. Here's my backup: Why did the scarecrow win an award? Because he was outstanding in his field!"
)

func (h BotCommandHelp) Joke(request BotCommand) (response HelpResponse) {
	response.Help = `Tells a joke from reddit, an AI joke generator, or a backup joke feed`
	response.Description = `Bennder delivers original dad jokes with sarcastic robot confidence.`

	return response
}

func (bc BotCommand) Joke(event BotCommand) error {
	u, ok, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if !ok {
		return err
	}
	response := comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		Message:        "",
		Type:           "command",
		UserId:         u.Id,
	}
	hc := &http.Client{Timeout: 10 * time.Second}
	providers := jokes.Composite{
		OnFailure: func(provider string, err error) {
			log.Printf("Joke provider %s failed: %v", provider, err)
		},
		Providers: []jokes.Provider{jokes.Reddit{
			Client:       hc,
			TokenURL:     redditTokenURL,
			FeedURL:      redditDadJokesURL,
			ClientID:     redditClientID,
			ClientSecret: redditSecret,
			UserAgent:    redditUserAgent,
		}},
	}
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		providers.Providers = append(providers.Providers, jokes.OpenAI{
			Client: &http.Client{Timeout: 30 * time.Second},
			APIKey: apiKey,
			Model:  os.Getenv("OPENAI_MODEL"),
		})
	}
	providers.Providers = append(providers.Providers, jokes.DadJokeDaily{Client: hc})
	return fetchJoke(event, response, providers)
}

func fetchJoke(event BotCommand, response comms.Response, provider jokes.Provider) error {
	joke, err := provider.Random(context.Background())
	if err != nil {
		return jokeRequestFailed(event, response, err)
	}

	response.Type = "post"
	response.Message = joke.Setup
	if response.Message == "" {
		response.Message = joke.Punchline
		joke.Punchline = ""
	}
	event.ResponseChannel <- response
	if joke.Punchline == "" {
		return nil
	}

	response.Type = "command"
	response.Message = fmt.Sprintf(`/echo "%s" 5`, joke.Punchline)
	event.ResponseChannel <- response
	return nil
}

func jokeRequestFailed(event BotCommand, response comms.Response, err error) error {
	log.Printf("Joke command: %v", err)
	response.Type = "dm"
	response.Message = "I couldn't fetch a joke right now. Please try again later."
	event.ResponseChannel <- response

	response.Type = "post"
	response.Message = dadJokeFallback
	event.ResponseChannel <- response
	return nil
}
