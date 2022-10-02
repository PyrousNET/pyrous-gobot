package users

import (
	"encoding/json"
	"fmt"
	"reflect"

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
		RpsPlaying    string `json:"rps-playing"`
		FarkleValue   int    `json:"farkle-value"`
		FarklePlaying string `json:"farkle-playing"`
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
				RpsPlaying:    "",
				FarklePlaying: "",
				FarkleValue:   0,
				SyndFeed:      "",
				FeedCount:     0,
			}
			u, _ := json.Marshal(newUser)
			c.Put(key, u)
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

	u, ok, err := c.Get(key)
	if err != nil {
		return User{}, false, fmt.Errorf("error communicating with redis: %v", err)
	}

	if ok {
		var user User
		user, err = getUserFromUnknownType(u, user, err)
		if err != nil {
			return User{}, false, fmt.Errorf("error unmarshalling user: %v", err)
		}

		return user, ok, nil
	}

	return User{}, ok, nil
}

func GetUsers(c cache.Cache) ([]User, bool, error) {
	userKeys := c.GetKeys(KeyPrefix)
	var users []User
	var err error

	for _, key := range userKeys {
		u, ok, e := c.Get(key)
		err = e

		if ok {
			var user User
			user, err = getUserFromUnknownType(u, user, err)
			users = append(users, user)
		} else {
			break
		}
	}

	return users, true, err
}

func getUserFromUnknownType(u interface{}, user User, err error) (User, error) {
	if reflect.TypeOf(u).String() != "[]uint8" {
		var user User
		var jm map[string]interface{}
		err = json.Unmarshal([]byte(u.(string)), &jm)
		jb, _ := json.Marshal(jm)
		err = json.Unmarshal(jb, &user)
	} else {
		err = json.Unmarshal(u.([]byte), &user)
	}
	return user, err
}

func UpdateUser(user User, c cache.Cache) (User, bool) {
	key := KeyPrefix + user.Name

	c.Put(key, user)

	return user, true
}
