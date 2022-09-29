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
	thoughtFeed struct {
		Links []thoughtLink `json:"links,omitempty"`
	}

	thoughtLink struct {
		Title    string `json:"title,omitempty"`
		Over18   bool   `json:"over_18,omitempty"`
		Stickied bool   `json:"stickied,omitempty"`
	}
)

func (h BotCommandHelp) Thought(request BotCommand) (response HelpResponse) {
	response.Help = `Have Bender give a random "shower-thought"`
	response.Description = `Have Bender give a random "shower-thought"`

	return response
}

func (bc BotCommand) Thought(event BotCommand) (response Response, err error) {
	url := "https://www.teddit.net/r/Showerthoughts/?api"
	hc := &http.Client{Timeout: 10 * time.Second}
	r, err := hc.Get(url)
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

	var feed thoughtFeed
	err = json.Unmarshal(b, &feed)
	if err != nil {
		log.Fatal(err)
	}

	response.Type = "post"

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(feed.Links), func(i, j int) { feed.Links[i], feed.Links[j] = feed.Links[j], feed.Links[i] })

	for _, link := range feed.Links {
		if !link.Over18 && !link.Stickied {
			response.Message = link.Title
			return response, nil
		}
	}

	response.Message = "I couldn't find anything that wouldn't make you blush. :-("
	return response, nil
}
