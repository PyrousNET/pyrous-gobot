package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const redditUserAgent = "go:github.com/PyrousNET/pyrous-gobot:v1.0.0"

type (
	jokeFeed struct {
		Data struct {
			Children []struct {
				JokeData jokeData `json:"data,omitempty"`
			} `json:"children,omitempty"`
		} `json:"data,omitempty"`
	}

	jokeData struct {
		Title        string                   `json:"title,omitempty"`
		Over18       bool                     `json:"over_18,omitempty"`
		Stickied     bool                     `json:"stickied,omitempty"`
		Selftext     string                   `json:"selftext,omitempty"`
		IsVideo      bool                     `json:"is_video,omitempty"`
		AllAwardings []map[string]interface{} `json:"all_awardings,omitempty"`
	}

	authToken struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Expires     int32  `json:"expires_in"`
		Scope       string `json:"scope"`
	}
)

func (h BotCommandHelp) Joke(request BotCommand) (response HelpResponse) {
	response.Help = `Pulls random dad jokes from reddit r/DadJokes`
	response.Description = `Bennder tells those sweet dad jokes like no one else.`

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
	token_uri := "https://www.reddit.com/api/v1/access_token"
	uri := "https://oauth.reddit.com/r/dadjokes"
	hc := &http.Client{Timeout: 10 * time.Second}
	return fetchJoke(event, response, hc, token_uri, uri)
}

func fetchJoke(event BotCommand, response comms.Response, hc *http.Client, token_uri, uri string) error {
	//Get Reddit Access Token
	req, err := http.NewRequest("POST", token_uri, strings.NewReader("grant_type=client_credentials"))
	req.SetBasicAuth("aIuZxRUiUiPIFD-fVb--jg", "UpGXB262RUsADk1RNU3vaMqLFCKxmQ")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", redditUserAgent)
	r, err := hc.Do(req)
	if err != nil {
		response.Type = "dm"
		response.Message = "Failed to get reddit access token: " + err.Error()
		event.ResponseChannel <- response
		return err
	}
	defer r.Body.Close()

	if r.StatusCode < http.StatusOK || r.StatusCode >= http.StatusMultipleChoices {
		return jokeRequestFailed(event, response, redditHTTPError("token request", r))
	}
	b, err := io.ReadAll(r.Body)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}

	var auth authToken
	err = json.Unmarshal(b, &auth)
	if err != nil || auth.AccessToken == "" {
		if err == nil {
			err = fmt.Errorf("reddit token response did not contain an access token")
		}
		return jokeRequestFailed(event, response, fmt.Errorf("could not parse reddit token response: %w", err))
	}

	bearer := "Bearer " + auth.AccessToken

	// Get Jokes List
	req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}
	req.Header.Add("Authorization", bearer)
	req.Header.Set("User-Agent", redditUserAgent)

	r, err = hc.Do(req)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}
	defer r.Body.Close()

	if r.StatusCode < http.StatusOK || r.StatusCode >= http.StatusMultipleChoices {
		return jokeRequestFailed(event, response, redditHTTPError("joke request", r))
	}
	b, err = io.ReadAll(r.Body)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}

	var feed jokeFeed
	err = json.Unmarshal(b, &feed)
	if err != nil {
		return jokeRequestFailed(event, response, fmt.Errorf("could not parse reddit joke response: %w", err))
	}
	if len(feed.Data.Children) == 0 {
		return jokeRequestFailed(event, response, fmt.Errorf("reddit returned no jokes"))
	}

	response.Type = "post"

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(feed.Data.Children), func(i, j int) {
		feed.Data.Children[i], feed.Data.Children[j] = feed.Data.Children[j], feed.Data.Children[i]
	})

	foundContent := false
	for _, child := range feed.Data.Children {
		jokeData := child.JokeData
		title := strings.TrimSpace(jokeData.Title)
		selftext := strings.TrimSpace(jokeData.Selftext)
		if title == "" && selftext == "" {
			continue
		}
		foundContent = true

		if !jokeData.Over18 && !jokeData.Stickied && !jokeData.IsVideo {
			response.Message = jokeData.Title
			if title == "" {
				response.Message = jokeData.Selftext
			}
			event.ResponseChannel <- response
			if selftext == "" {
				return nil
			}
			response.Type = "command"
			response.Message = "/echo \"" + jokeData.Selftext + "\" 5"
			event.ResponseChannel <- response
			return nil
		}
	}

	if !foundContent {
		return jokeRequestFailed(event, response, fmt.Errorf("reddit returned no usable jokes"))
	}

	response.Message = "I couldn't find anything that wouldn't make you blush. :-("
	event.ResponseChannel <- response
	return nil
}

func jokeRequestFailed(event BotCommand, response comms.Response, err error) error {
	log.Printf("Joke command: %v", err)
	response.Type = "dm"
	response.Message = "I couldn't fetch a joke right now. Please try again later."
	event.ResponseChannel <- response

	response.Type = "post"
	response.Message = "Reddit's joke drawer is locked, so here's one from mine: Why did the scarecrow win an award? Because he was outstanding in his field!"
	event.ResponseChannel <- response
	return nil
}

func redditHTTPError(operation string, response *http.Response) error {
	const maxErrorBodyLength = 512
	body, err := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyLength+1))
	detail := strings.TrimSpace(string(body))
	if len(detail) > maxErrorBodyLength {
		detail = detail[:maxErrorBodyLength] + "..."
	}
	if err != nil {
		return fmt.Errorf("reddit %s returned HTTP %s (could not read error body: %v)", operation, response.Status, err)
	}
	if detail == "" {
		return fmt.Errorf("reddit %s returned HTTP %s", operation, response.Status)
	}
	return fmt.Errorf("reddit %s returned HTTP %s: %q", operation, response.Status, detail)
}
