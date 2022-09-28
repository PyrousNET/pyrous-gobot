package users

import (
	"fmt"
	"github.com/pyrousnet/mattermost-golang-bot/internal/mmclient"
)

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

func SetupUsers(mm *mmclient.MMClient) ([]User, error) {
	users, r := mm.Client.GetKnownUsers()
	if r.StatusCode == 200 {
		for _, u := range users {
			user, _ := mm.Client.GetUser(u, "")
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
			// TODO persist these to Redis
			fmt.Printf("%v", newUser)
		}
	}
	return nil, nil
}
