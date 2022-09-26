package main

type User struct {
	Name          string `json:"name"`
	Message       string `json:"message"`
	Rps           string `json:"rps"`
	RpsPlaying    bool   `json:"rps-playing"`
	FarkleValue   int    `json:"farkle-value"`
	FarklePlaying bool   `json:"farkle-playing"`
	SyndFeed      string `json:"synd-feed"`
	FeedCount     int    `json:"feed-count"`
}
