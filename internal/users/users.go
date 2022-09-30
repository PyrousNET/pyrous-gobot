package users

import (
	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/mmclient"
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
	userIds, r, err := mm.Client.GetKnownUsers()
	if r.StatusCode == 200 {
		for _, u := range userIds {
			user, _, _ := mm.Client.GetUser(u, "")
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
	return err
}

func HandlePost(post *model.Post, mm *mmclient.MMClient, c cache.Cache) error {
	user, _, err := mm.Client.GetUser(post.UserId, "")
	key := KeyPrefix + user.Username
	persisted, ok, _ := GetUser(user.Username, c)

	if ok {
		persisted.Message = post.Message
		c.Put(key, persisted)
	}

	return err
}

func GetUser(username string, c cache.Cache) (User, bool, error) {
	key := KeyPrefix + username
	user, ok, err := c.Get(key)
	if err != nil {
		return User{}, false, err
	}

	if ok {
		return user.(User), ok, nil
	}

	return User{}, ok, nil
}
