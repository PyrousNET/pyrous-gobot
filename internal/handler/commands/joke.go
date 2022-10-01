package commands

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"
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
)

func (h BotCommandHelp) Joke(request BotCommand) (response HelpResponse) {
	response.Help = `Pulls random dad jokes from reddit r/DadJokes`
	response.Description = `Bennder tells those sweet dad jokes like no one else.`

	return response
}

func (bc BotCommand) Joke(event BotCommand) (response Response, err error) {
	uri := "https://www.teddit.net/r/dadjokes/?api&target=reddit"
	hc := &http.Client{Timeout: 10 * time.Second}
	r, err := hc.Get(uri)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		return response, err
	}
	defer r.Body.Close()

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		response.Type = "dm"
		response.Message = err.Error()
		return response, err
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
			response.Delay = time.Duration(5 * time.Second)
			response.Message2 = jokeData.Selftext
			return response, nil
		}
	}

	response.Message = "I couldn't find anything that wouldn't make you blush. :-("
	return response, nil
}
