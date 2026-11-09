package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/jokes"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const (
	jokeFeedbackPendingPrefix = "joke-feedback:pending:"
	jokeFeedbackRatedPrefix   = "joke-feedback:rated:"
	jokeFeedbackPendingTTL    = 24 * time.Hour
	defaultJokeFeedbackCSV    = "joke-feedback.csv"
)

type pendingJokeFeedback struct {
	ID   string     `json:"id"`
	Joke jokes.Joke `json:"joke"`
}

var jokeFeedbackMu sync.Mutex

func (BotCommandHelp) Jokefeedback(BotCommand) (response HelpResponse) {
	response.Help = "Rate your latest joke from 1 (nonsense) to 5 (great), with an optional comment. Usage: '!jokefeedback <1-5> [comment]'"
	response.Description = "Rate the latest joke"
	return response
}

func (bc BotCommand) Jokefeedback(event BotCommand) error {
	fields := strings.Fields(event.body)
	if len(fields) == 0 {
		postJokeFeedbackResponse(event, "Rate your latest joke from 1 (nonsense) to 5 (great) with `!jokefeedback <1-5> [comment]`.")
		return nil
	}
	rating, err := strconv.Atoi(fields[0])
	if err != nil || rating < 1 || rating > 5 {
		postJokeFeedbackResponse(event, "Please use a rating from 1 to 5, optionally followed by a comment.")
		return nil
	}
	if event.cache == nil {
		postJokeFeedbackResponse(event, "I couldn't find a recent joke in this channel to rate.")
		return nil
	}

	user, ok, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil || !ok {
		postJokeFeedbackResponse(event, "I couldn't find your account to save that rating.")
		return nil
	}
	key := pendingJokeFeedbackKey(event.ReplyChannel.Id)
	value, ok, err := event.cache.Get(key)
	if err != nil || !ok {
		postJokeFeedbackResponse(event, "I couldn't find a recent joke in this channel to rate.")
		return nil
	}
	serialized, ok := value.(string)
	if !ok {
		postJokeFeedbackResponse(event, "I couldn't find a recent joke in this channel to rate.")
		return nil
	}
	var pending pendingJokeFeedback
	if err := json.Unmarshal([]byte(serialized), &pending); err != nil {
		postJokeFeedbackResponse(event, "I couldn't find a recent joke in this channel to rate.")
		return nil
	}

	jokeFeedbackMu.Lock()
	defer jokeFeedbackMu.Unlock()
	ratedKey := jokeFeedbackRatedKey(event.ReplyChannel.Id, pending.ID, user.Id)
	if _, alreadyRated, err := event.cache.Get(ratedKey); err != nil {
		postJokeFeedbackResponse(event, "I couldn't check whether you've already rated this joke. Please try again.")
		return nil
	} else if alreadyRated {
		postJokeFeedbackResponse(event, "You've already rated this joke.")
		return nil
	}

	comment := strings.Join(fields[1:], " ")
	path := jokeFeedbackCSVPath()
	if err := jokes.AppendFeedback(path, jokes.JokeFeedback{
		Timestamp: time.Now().UTC(),
		UserID:    user.Id,
		JokeID:    pending.ID,
		Rating:    rating,
		Comment:   comment,
		Joke:      pending.Joke,
	}); err != nil {
		log.Printf("Unable to save joke feedback: %v", err)
		postJokeFeedbackResponse(event, "I couldn't save that rating right now. Please try again later.")
		return nil
	}
	updateCachedJokeFeedbackProfile(event.cache, path, rating, comment)
	event.cache.Put(ratedKey, "true")
	if expirer, ok := event.cache.(interface{ Expire(string, time.Duration) }); ok {
		expirer.Expire(ratedKey, jokeFeedbackPendingTTL)
	}
	postJokeFeedbackResponse(event, fmt.Sprintf("Thanks! I saved your %d/5 rating.", rating))
	return nil
}

func rememberJokeForFeedback(c cache.Cache, channelID string, joke jokes.Joke) error {
	if c == nil {
		return nil
	}
	data, err := json.Marshal(pendingJokeFeedback{ID: uuid.NewString(), Joke: joke})
	if err != nil {
		return err
	}
	key := pendingJokeFeedbackKey(channelID)
	c.Put(key, string(data))
	if expirer, ok := c.(interface{ Expire(string, time.Duration) }); ok {
		expirer.Expire(key, jokeFeedbackPendingTTL)
	}
	return nil
}

func pendingJokeFeedbackKey(channelID string) string {
	return jokeFeedbackPendingPrefix + channelID
}

func jokeFeedbackCSVPath() string {
	if path := os.Getenv("JOKE_FEEDBACK_CSV"); path != "" {
		return path
	}
	return defaultJokeFeedbackCSV
}

func cachedJokeFeedbackGuidance(c cache.Cache, path string) string {
	if c == nil {
		return ""
	}
	if value, ok, err := c.Get(jokes.JokeFeedbackProfileCacheKey); err == nil && ok {
		if serialized, ok := value.(string); ok {
			var profile jokes.JokeFeedbackProfile
			if json.Unmarshal([]byte(serialized), &profile) == nil {
				return profile.GenerationGuidance()
			}
		}
	}
	profile, err := jokes.LoadJokeFeedbackProfile(path)
	if err != nil {
		log.Printf("Unable to load joke feedback profile: %v", err)
		return ""
	}
	serialized, err := json.Marshal(profile)
	if err != nil {
		log.Printf("Unable to cache joke feedback profile: %v", err)
		return ""
	}
	c.Put(jokes.JokeFeedbackProfileCacheKey, string(serialized))
	return profile.GenerationGuidance()
}

func updateCachedJokeFeedbackProfile(c cache.Cache, path string, rating int, comment string) {
	if c == nil {
		return
	}
	var profile jokes.JokeFeedbackProfile
	value, found, cacheErr := c.Get(jokes.JokeFeedbackProfileCacheKey)
	if cacheErr == nil && found {
		if serialized, ok := value.(string); ok && json.Unmarshal([]byte(serialized), &profile) == nil {
			profile.Add(rating, comment)
			storeJokeFeedbackProfile(c, profile)
			return
		}
	}
	var err error
	profile, err = jokes.LoadJokeFeedbackProfile(path)
	if err != nil {
		log.Printf("Unable to load joke feedback profile: %v", err)
		return
	}
	storeJokeFeedbackProfile(c, profile)
}

func storeJokeFeedbackProfile(c cache.Cache, profile jokes.JokeFeedbackProfile) {
	serialized, err := json.Marshal(profile)
	if err != nil {
		log.Printf("Unable to cache joke feedback profile: %v", err)
		return
	}
	c.Put(jokes.JokeFeedbackProfileCacheKey, string(serialized))
}

func jokeFeedbackRatedKey(channelID, jokeID, userID string) string {
	return jokeFeedbackRatedPrefix + channelID + ":" + jokeID + ":" + userID
}

func postJokeFeedbackResponse(event BotCommand, message string) {
	var user users.User
	if event.cache != nil {
		user, _, _ = users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	}
	event.ResponseChannel <- comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         user.Id,
		Type:           "post",
		Message:        message,
	}
}
