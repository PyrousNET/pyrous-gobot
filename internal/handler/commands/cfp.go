package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const (
	cfpRankingsURL    = "https://collegefootballplayoff.com/rankings"
	cfbdAPIBaseURL    = "https://api.collegefootballdata.com"
	cfpCacheKeyPrefix = "cfp-rankings:v2:"
)

type cfpRank struct {
	Rank     int    `json:"rank"`
	TeamID   int    `json:"teamId"`
	TeamName string `json:"school"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Ties     int    `json:"ties"`
	Movement int    `json:"movement"`
}

type cfbdPoll struct {
	Poll  string    `json:"poll"`
	Ranks []cfpRank `json:"ranks"`
}

type cfbdRankingSnapshot struct {
	Week  int        `json:"week"`
	Polls []cfbdPoll `json:"polls"`
}

type cfpRankingCache struct {
	Week  string    `json:"week"`
	Ranks []cfpRank `json:"ranks"`
}

type cfbdTeamRecord struct {
	TeamID int `json:"teamId"`
	Total  struct {
		Wins   int `json:"wins"`
		Losses int `json:"losses"`
		Ties   int `json:"ties"`
	} `json:"total"`
}

func (BotCommandHelp) Cfp(BotCommand) (response HelpResponse) {
	response.Help = "Fetch the latest College Football Playoff committee rankings. Usage: '!cfp'. Requires CFBD_API_KEY."
	response.Description = "Get the latest CFP rankings"
	return response
}

func (bc BotCommand) Cfp(event BotCommand) error {
	return bc.cfpWithFetcher(event, fetchCachedCFPRankings)
}

func (bc BotCommand) cfpWithFetcher(event BotCommand, fetcher func(*http.Client, cache.Cache) ([]cfpRank, string, error)) error {
	u, ok, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("could not find user %s", event.sender)
	}

	event.ResponseChannel <- comms.Response{ReplyChannelId: event.ReplyChannel.Id, UserId: u.Id, Type: "post", Message: "I'm pulling the latest CFP rankings now—please wait a moment."}
	ranks, week, err := fetcher(http.DefaultClient, event.cache)
	if err != nil {
		return fmt.Errorf("unable to fetch CFP rankings: %w", err)
	}
	event.ResponseChannel <- comms.Response{ReplyChannelId: event.ReplyChannel.Id, UserId: u.Id, Type: "post", Message: formatCFPRankings(ranks, week)}
	return nil
}

func fetchCachedCFPRankings(client *http.Client, c cache.Cache) ([]cfpRank, string, error) {
	key := os.Getenv("CFBD_API_KEY")
	if key == "" {
		return nil, "", fmt.Errorf("CFBD_API_KEY is not configured")
	}
	year := time.Now().Year()
	return fetchCachedCFPRankingsWithFetcher(client, c, key, year, fetchCFPRankings)
}

func fetchCachedCFPRankingsWithFetcher(client *http.Client, c cache.Cache, key string, year int, fetcher func(*http.Client, string, int) ([]cfpRank, string, error)) ([]cfpRank, string, error) {
	cacheKey := fmt.Sprintf("%s%d", cfpCacheKeyPrefix, year)
	if c != nil {
		if value, ok, err := c.Get(cacheKey); err == nil && ok {
			var result cfpRankingCache
			if data, ok := value.(string); ok && json.Unmarshal([]byte(data), &result) == nil && len(result.Ranks) > 0 {
				return result.Ranks, result.Week, nil
			}
		}
	}
	ranks, week, err := fetcher(client, key, year)
	if err != nil {
		return nil, "", err
	}
	if c != nil {
		serialized, marshalErr := json.Marshal(cfpRankingCache{Week: week, Ranks: ranks})
		if marshalErr == nil {
			c.Put(cacheKey, string(serialized))
			if expirer, ok := c.(interface{ Expire(string, time.Duration) }); ok {
				expirer.Expire(cacheKey, durationUntilNextCFPRankings(time.Now()))
			}
		}
	}
	return ranks, week, nil
}

func durationUntilNextCFPRankings(now time.Time) time.Duration {
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		location = time.FixedZone("MST", -7*60*60)
	}
	localNow := now.In(location)
	daysUntilTuesday := (int(time.Tuesday) - int(localNow.Weekday()) + 7) % 7
	next := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+daysUntilTuesday, 19, 0, 0, 0, location)
	if !next.After(localNow) {
		next = next.AddDate(0, 0, 7)
	}
	return next.Sub(localNow)
}

func fetchCFPRankings(client *http.Client, key string, year int) ([]cfpRank, string, error) {
	return fetchCFPRankingsFromURL(client, cfbdAPIBaseURL, key, year)
}

func fetchCFPRankingsFromURL(client *http.Client, apiBaseURL, key string, year int) ([]cfpRank, string, error) {
	rankingsURL, err := cfbdEndpointURL(apiBaseURL, "/rankings", url.Values{"year": {fmt.Sprint(year)}, "poll": {"cfp"}})
	if err != nil {
		return nil, "", err
	}
	rankingsBody, err := fetchCFBDResponse(client, rankingsURL, key)
	if err != nil {
		return nil, "", err
	}
	var snapshots []cfbdRankingSnapshot
	if err := json.Unmarshal(rankingsBody, &snapshots); err != nil {
		return nil, "", fmt.Errorf("decode CFBD rankings: %w", err)
	}
	if len(snapshots) == 0 {
		return nil, "", fmt.Errorf("CFBD has not published CFP rankings for %d", year)
	}

	latestIndex := 0
	previousIndex := -1
	for i := range snapshots {
		if snapshots[i].Week > snapshots[latestIndex].Week {
			latestIndex = i
		}
	}
	for i := range snapshots {
		if i != latestIndex && snapshots[i].Week < snapshots[latestIndex].Week &&
			(previousIndex == -1 || snapshots[i].Week > snapshots[previousIndex].Week) {
			previousIndex = i
		}
	}
	latest := snapshots[latestIndex]
	ranks := cfpRanksForPoll(latest.Polls)
	if len(ranks) == 0 {
		return nil, "", fmt.Errorf("CFBD returned no CFP rankings for week %d", latest.Week)
	}
	previousRanks := make(map[int]int)
	if previousIndex >= 0 {
		for _, rank := range cfpRanksForPoll(snapshots[previousIndex].Polls) {
			previousRanks[rank.TeamID] = rank.Rank
		}
	}

	recordsURL, err := cfbdEndpointURL(apiBaseURL, "/records", url.Values{"year": {fmt.Sprint(year)}})
	if err != nil {
		return nil, "", err
	}
	recordsBody, err := fetchCFBDResponse(client, recordsURL, key)
	if err != nil {
		return nil, "", err
	}
	var records []cfbdTeamRecord
	if err := json.Unmarshal(recordsBody, &records); err != nil {
		return nil, "", fmt.Errorf("decode CFBD team records: %w", err)
	}
	recordsByTeamID := make(map[int]cfbdTeamRecord, len(records))
	for _, record := range records {
		recordsByTeamID[record.TeamID] = record
	}
	for i := range ranks {
		if record, ok := recordsByTeamID[ranks[i].TeamID]; ok {
			ranks[i].Wins = record.Total.Wins
			ranks[i].Losses = record.Total.Losses
			ranks[i].Ties = record.Total.Ties
		}
		if previousRank, ok := previousRanks[ranks[i].TeamID]; ok {
			ranks[i].Movement = previousRank - ranks[i].Rank
		}
	}
	sort.Slice(ranks, func(i, j int) bool { return ranks[i].Rank < ranks[j].Rank })
	return ranks, fmt.Sprintf("Week %d", latest.Week), nil
}

func cfpRanksForPoll(polls []cfbdPoll) []cfpRank {
	for _, poll := range polls {
		if strings.Contains(strings.ToLower(poll.Poll), "playoff") || strings.EqualFold(poll.Poll, "cfp") {
			return poll.Ranks
		}
	}
	if len(polls) == 1 {
		return polls[0].Ranks
	}
	return nil
}

func cfbdEndpointURL(apiBaseURL, path string, query url.Values) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(apiBaseURL, "/") + path)
	if err != nil {
		return "", err
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func fetchCFBDResponse(client *http.Client, endpoint, key string) ([]byte, error) {
	if client == nil {
		client = http.DefaultClient
	}
	request, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+key)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("CFBD returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func formatCFPRankings(ranks []cfpRank, week string) string {
	lines := []string{fmt.Sprintf("### CFP Rankings (%s)", week), "| Rank | Team | Record | Movement |", "| ---: | --- | :---: | :---: |"}
	for _, rank := range ranks {
		lines = append(lines, fmt.Sprintf("| %d | %s | %s | %s |", rank.Rank, rank.TeamName, formatCFPRecord(rank), formatAPTrend(rank.Movement)))
	}
	return strings.Join(append(lines, "", fmt.Sprintf("Source: [College Football Playoff](%s)", cfpRankingsURL)), "\n")
}

func formatCFPRecord(rank cfpRank) string {
	return fmt.Sprintf("%d-%d-%d", rank.Wins, rank.Losses, rank.Ties)
}
