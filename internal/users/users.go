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
		Id        string `json:"id"`
		Name      string `json:"name"`
		Message   string `json:"message"`
		SyndFeed  string `json:"synd-feed"`
		FeedCount int    `json:"feed-count"`
	}
)

func SetupUsers(mm *mmclient.MMClient, c cache.Cache) error {
	userIds, r, err := mm.Client.GetKnownUsers()
	if r.StatusCode == 200 {
		for _, u := range userIds {
			user, _, _ := mm.Client.GetUser(u, "")
			key := KeyPrefix + user.Username
			newUser := User{
				Id:        u,
				Name:      user.Username,
				Message:   "",
				SyndFeed:  "",
				FeedCount: 0,
			}
			u, _ := json.Marshal(newUser)
			c.Put(key, u)
		}
	}
	return err
}

func HandlePost(post *model.Post, mm *mmclient.MMClient, c cache.Cache) error {
	user, _, err := mm.Client.GetUser(post.UserId, "")
	if err == nil {
		key := KeyPrefix + user.Username
		persisted, ok, _ := GetUser(user.Username, c)

		if ok {
			persisted.Message = post.Message

			ub, _ := json.Marshal(persisted)
			c.Put(key, ub)
		}
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
		usr, err := getUserFromUnknownType(u, err)
		if err != nil {
			return User{}, false, fmt.Errorf("error unmarshalling user: %v", err)
		}

		return usr, ok, nil
	}

	return User{}, ok, nil
}

func GetUsers(c cache.Cache) ([]User, bool, error) {
	userKeys, kerr := c.GetKeys(KeyPrefix)
	if kerr != nil {
		return []User{}, false, kerr
	}
	var users []User
	var err error

	for _, key := range userKeys {
		u, ok, e := c.Get(key)
		err = e

		if ok {
			var user User
			user, err = getUserFromUnknownType(u, err)
			users = append(users, user)
		} else {
			break
		}
	}

	return users, true, err
}

func getUserFromUnknownType(u interface{}, err error) (User, error) {
	var user User

	if reflect.TypeOf(u).String() != "[]uint8" {
		err := json.Unmarshal([]byte(u.(string)), &user)
		return user, err
	}

	err = json.Unmarshal(u.([]byte), &user)

	return user, err
}

func UpdateUser(user User, c cache.Cache) (User, bool) {
	key := KeyPrefix + user.Name

	c.Put(key, user)

	return user, true
}
