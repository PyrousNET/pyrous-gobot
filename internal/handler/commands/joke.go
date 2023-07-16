package commands

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

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

	//Get Reddit Access Token
	req, err := http.NewRequest("POST", token_uri, strings.NewReader("grant_type=client_credentials"))
	req.SetBasicAuth("aIuZxRUiUiPIFD-fVb--jg", "UpGXB262RUsADk1RNU3vaMqLFCKxmQ")
	r, err := hc.Do(req)
	if err != nil {
		response.Type = "dm"
		response.Message = "Failed to get reddit access token: " + err.Error()
		event.ResponseChannel <- response
		return err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}

	var auth authToken
	err = json.Unmarshal(b, &auth)
	if err != nil {
		log.Fatal(err)
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

	r, err = hc.Do(req)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}
	defer r.Body.Close()

	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		event.ResponseChannel <- response
		return err
	}

	var feed jokeFeed
	err = json.Unmarshal(b, &feed)
	if err != nil {
		log.Fatal(err)
	}

	response.Type = "post"

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(feed.Data.Children), func(i, j int) {
		feed.Data.Children[i], feed.Data.Children[j] = feed.Data.Children[j], feed.Data.Children[i]
	})

	for _, child := range feed.Data.Children {
		jokeData := child.JokeData
		if !jokeData.Over18 && !jokeData.Stickied && !jokeData.IsVideo {
			response.Message = jokeData.Title
			event.ResponseChannel <- response
			response.Type = "command"
			response.Message = "/echo \"" + jokeData.Selftext + "\" 5"
			event.ResponseChannel <- response
			return nil
		}
	}

	response.Message = "I couldn't find anything that wouldn't make you blush. :-("
	event.ResponseChannel <- response
	return nil
}
