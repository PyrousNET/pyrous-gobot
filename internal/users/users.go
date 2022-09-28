package users

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/pyrousnet/mattermost-golang-bot/internal/cache"
	"github.com/pyrousnet/mattermost-golang-bot/internal/mmclient"
)

const KeyPrefix = "user-"

type (
	User struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		Message       string `json:"message"`
		Rps           string `json:"rps"`
		RpsPlaying    bool   `json:"rps-playing"`
		FarkleValue   int    `json:"farkle-value"`
		FarklePlaying bool   `json:"farkle-playing"`
		SyndFeed      string `json:"synd-feed"`
		FeedCount     int    `json:"feed-count"`
	}
)

func SetupUsers(mm *mmclient.MMClient, c cache.Cache) error {
	userIds, r := mm.Client.GetKnownUsers()
	if r.StatusCode == 200 {
		for _, u := range userIds {
			user, _ := mm.Client.GetUser(u, "")
			key := KeyPrefix + user.Username
			newUser := User{
				Id:            u,
				Name:          user.Username,
				Message:       "",
				Rps:           "",
				RpsPlaying:    false,
				FarklePlaying: false,
				FarkleValue:   0,
				SyndFeed:      "",
				FeedCount:     0,
			}
			c.Put(key, newUser)
		}
	}
	return nil
}

func HandlePost(post *model.Post, mm *mmclient.MMClient, c cache.Cache) error {
	user, _ := mm.Client.GetUser(post.UserId, "")
	key := KeyPrefix + user.Username
	persisted, _ := GetUser(user.Username, c)

	persisted.Message = post.Message
	c.Put(key, persisted)

	return nil
}

func GetUser(username string, c cache.Cache) (User, error) {
	key := KeyPrefix + username
	user := c.Get(key)
	return user.(User), nil
}
