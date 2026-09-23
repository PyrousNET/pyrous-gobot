package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const apTop25URL = "https://apnews.com/hub/ap-top-25-college-football-poll"

const (
	apTop25CacheKey = "ap-top25:current"
	apTop25CacheTTL = time.Hour
)

const espnNCAAFScoreboardURL = "https://site.api.espn.com/apis/site/v2/sports/football/college-football/scoreboard"

const (
	espnScoreboardAttempts = 3
	espnRetryDelay         = 250 * time.Millisecond
)

// AP_TOP_25_POLL_ID can override the current season's poll ID. This fallback
// lets the bot use AP's rankings API when Cloudflare blocks the HTML page.
const apTop25FallbackPollID = "0000019e-d6c2-d6a8-a59e-feea16ca0000"

var (
	apPollIDPattern     = regexp.MustCompile(`data-top25-poll-id\s*=\s*"([^"]+)"`)
	apWeekPattern       = regexp.MustCompile(`data-week\s*=\s*"([^"]+)"`)
	apWeekOptionPattern = regexp.MustCompile(`(?s)<option\s+value="([^"]+)"\s+data-close-at=`)
	apAPIBasePattern    = regexp.MustCompile(`data-api-base-url\s*=\s*"([^"]+)"`)
)

type apTop25Rank struct {
	Rank     int    `json:"rank"`
	Trend    int    `json:"trend"`
	TeamName string `json:"teamName"`
	Wins     int    `json:"win"`
	Losses   int    `json:"loss"`
	Ties     int    `json:"tie"`
}

type apTop25Result struct {
	Week  string        `json:"week"`
	Ranks []apTop25Rank `json:"ranks"`
}

type espnNCAAFScoreboard struct {
	Events []struct {
		Week struct {
			Number int `json:"number"`
		} `json:"week"`
	} `json:"events"`
}

func (h BotCommandHelp) Top25(request BotCommand) (response HelpResponse) {
	response.Help = "Fetch the latest AP Top 25 college football rankings. Usage: '!top25'."
	response.Description = "Get the latest AP Top 25 rankings"
	return response
}

func (bc BotCommand) Top25(event BotCommand) error {
	u, ok, err := users.GetUser(strings.TrimLeft(event.sender, "@"), event.cache)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("could not find user %s", event.sender)
	}

	teams, week, err := fetchCachedAPTop25(http.DefaultClient, event.cache)
	if err != nil {
		return fmt.Errorf("unable to fetch AP Top 25: %w", err)
	}

	event.ResponseChannel <- comms.Response{
		ReplyChannelId: event.ReplyChannel.Id,
		UserId:         u.Id,
		Type:           "post",
		Message:        formatAPTop25(teams, week),
	}
	return nil
}

// fetchCachedAPTop25 avoids repeatedly querying AP while still allowing the
// current poll to be discovered after a new ranking is published.
func fetchCachedAPTop25(client *http.Client, c cache.Cache) ([]apTop25Rank, string, error) {
	return fetchCachedAPTop25WithFetcher(client, c, fetchAPTop25)
}

func fetchCachedAPTop25WithFetcher(client *http.Client, c cache.Cache, fetcher func(*http.Client) ([]apTop25Rank, string, error)) ([]apTop25Rank, string, error) {
	if c != nil {
		if value, ok, err := c.Get(apTop25CacheKey); err == nil && ok {
			var result apTop25Result
			if data, ok := value.(string); ok && json.Unmarshal([]byte(data), &result) == nil && len(result.Ranks) > 0 {
				return result.Ranks, result.Week, nil
			}
		}
	}

	teams, week, err := fetcher(client)
	if err != nil {
		return nil, "", err
	}
	if c == nil {
		return teams, week, nil
	}

	data, err := json.Marshal(apTop25Result{Week: week, Ranks: teams})
	if err != nil {
		return teams, week, err
	}
	c.Put(apTop25CacheKey, string(data))
	if expiring, ok := c.(interface {
		Expire(string, time.Duration)
	}); ok {
		expiring.Expire(apTop25CacheKey, apTop25CacheTTL)
	}
	return teams, week, nil
}

func formatAPTop25(teams []apTop25Rank, week string) string {
	lines := make([]string, 0, len(teams)+5)
	lines = append(lines,
		fmt.Sprintf("### AP Top 25 (%s)", week),
		"| Rank | Team | Record | Movement |",
		"| ---: | --- | :---: | :---: |",
	)
	for _, team := range teams {
		lines = append(lines, fmt.Sprintf("| %d | %s | %s | %s |", team.Rank, team.TeamName, formatAPRecord(team), formatAPTrend(team.Trend)))
	}
	lines = append(lines, "", fmt.Sprintf("Source: [AP News — AP Top 25 Poll](%s)", apTop25URL))
	return strings.Join(lines, "\n")
}

func formatAPRecord(team apTop25Rank) string {
	return fmt.Sprintf("%d-%d-%d", team.Wins, team.Losses, team.Ties)
}

func formatAPTrend(trend int) string {
	switch {
	case trend > 0:
		return fmt.Sprintf("🟢 ▲ %d", trend)
	case trend < 0:
		return fmt.Sprintf("🔴 ▼ %d", -trend)
	default:
		return "—"
	}
}

func fetchAPTop25(client *http.Client) ([]apTop25Rank, string, error) {
	return fetchAPTop25FromURL(client, apTop25URL)
}

func fetchAPTop25FromURL(client *http.Client, pageURL string) ([]apTop25Rank, string, error) {
	if client == nil {
		client = http.DefaultClient
	}

	pageRequest, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, "", err
	}
	pageRequest.Header.Set("User-Agent", "pyrous-gobot/1.0")
	pageResponse, err := client.Do(pageRequest)
	if err != nil {
		return nil, "", err
	}
	defer pageResponse.Body.Close()
	if pageResponse.StatusCode != http.StatusOK {
		if pageResponse.StatusCode == http.StatusForbidden {
			return fetchAPTop25Fallback(client, pageURL)
		}
		return nil, "", fmt.Errorf("AP page returned HTTP %d", pageResponse.StatusCode)
	}

	page, err := io.ReadAll(io.LimitReader(pageResponse.Body, 10<<20))
	if err != nil {
		return nil, "", err
	}
	pollID := firstMatch(apPollIDPattern, page)
	week := firstMatch(apWeekPattern, page)
	apiBase := firstMatch(apAPIBasePattern, page)
	if pollID == "" || week == "" || apiBase == "" {
		return nil, "", fmt.Errorf("AP page did not contain poll metadata")
	}

	weeks := []string{week}
	for _, match := range apWeekOptionPattern.FindAllSubmatch(page, -1) {
		candidate := string(match[1])
		if candidate != "" && candidate != week {
			weeks = append(weeks, candidate)
		}
	}
	for _, candidate := range weeks {
		ranks, resultWeek, available, err := fetchAPTop25Week(client, apiBase, pollID, candidate, pageURL)
		if err != nil {
			return nil, "", err
		}
		if available {
			return ranks, resultWeek, nil
		}
	}
	return nil, "", fmt.Errorf("AP Top 25 is not published for the available weeks")
}

func fetchAPTop25Fallback(client *http.Client, pageURL string) ([]apTop25Rank, string, error) {
	pollID := os.Getenv("AP_TOP_25_POLL_ID")
	if pollID == "" {
		pollID = apTop25FallbackPollID
	}
	ncaafWeek, err := fetchCurrentNCAAFWeek(client, time.Now())
	if err != nil {
		return nil, "", fmt.Errorf("unable to determine current NCAA week: %w", err)
	}
	for _, week := range apWeeksForNCAAFWeek(ncaafWeek) {
		ranks, resultWeek, available, err := fetchAPTop25Week(client, "https://prod-api.apnews.com", pollID, week, pageURL)
		if err != nil {
			return nil, "", err
		}
		if available {
			return ranks, resultWeek, nil
		}
	}
	return nil, "", fmt.Errorf("AP page was blocked and no fallback poll was available")
}

func apWeeksForNCAAFWeek(ncaafWeek int) []string {
	if ncaafWeek <= 1 {
		return []string{"Preseason"}
	}
	weeks := make([]string, 0, ncaafWeek)
	for current := ncaafWeek - 1; current >= 1; current-- {
		weeks = append(weeks, fmt.Sprintf("Week %d", current))
	}
	weeks = append(weeks, "Preseason")
	return weeks
}

func fetchCurrentNCAAFWeek(client *http.Client, now time.Time) (int, error) {
	return fetchCurrentNCAAFWeekFromURL(client, now, espnNCAAFScoreboardURL)
}

func fetchCurrentNCAAFWeekFromURL(client *http.Client, now time.Time, endpoint string) (int, error) {
	if client == nil {
		client = http.DefaultClient
	}

	start := now.UTC().Truncate(24 * time.Hour)
	ranges := [][2]time.Time{
		{start, start.AddDate(0, 0, 6)},
		{start, start.AddDate(0, 0, 2)},
		{start.AddDate(0, 0, 3), start.AddDate(0, 0, 5)},
		{start.AddDate(0, 0, 6), start.AddDate(0, 0, 6)},
	}

	var lastErr error
	for _, dateRange := range ranges {
		week, err := fetchNCAAFWeekRange(client, endpoint, dateRange[0], dateRange[1])
		if err == nil {
			return week, nil
		}
		lastErr = err
		if !isRetryableESPNError(err) {
			return 0, err
		}
	}

	return 0, lastErr
}

func fetchNCAAFWeekRange(client *http.Client, endpoint string, start, end time.Time) (int, error) {
	queryURL, err := url.Parse(endpoint)
	if err != nil {
		return 0, err
	}
	query := queryURL.Query()
	query.Set("dates", start.Format("20060102")+"-"+end.Format("20060102"))
	queryURL.RawQuery = query.Encode()

	var lastErr error
	for attempt := 0; attempt < espnScoreboardAttempts; attempt++ {
		request, err := http.NewRequest(http.MethodGet, queryURL.String(), nil)
		if err != nil {
			return 0, err
		}
		response, err := client.Do(request)
		if err != nil {
			lastErr = fmt.Errorf("ESPN scoreboard request failed for %s: %w", queryURL.String(), err)
		} else {
			body, readErr := io.ReadAll(io.LimitReader(response.Body, 1<<20))
			response.Body.Close()
			if readErr != nil {
				return 0, readErr
			}
			if response.StatusCode != http.StatusOK {
				lastErr = fmt.Errorf("ESPN scoreboard returned HTTP %d for %s: %s", response.StatusCode, queryURL.String(), summarizeResponseBody(body))
				if !isRetryableESPNStatus(response.StatusCode) {
					return 0, lastErr
				}
			} else {
				var scoreboard espnNCAAFScoreboard
				if err := json.Unmarshal(body, &scoreboard); err != nil {
					return 0, err
				}
				for _, event := range scoreboard.Events {
					if event.Week.Number > 0 {
						return event.Week.Number, nil
					}
				}
				lastErr = fmt.Errorf("ESPN scoreboard returned no week for %s", queryURL.String())
			}
		}

		if attempt < espnScoreboardAttempts-1 {
			time.Sleep(espnRetryDelay * time.Duration(attempt+1))
		}
	}

	return 0, lastErr
}

func isRetryableESPNStatus(status int) bool {
	return status == http.StatusBadRequest || status == http.StatusForbidden || status == http.StatusTooManyRequests || status >= 500
}

func isRetryableESPNError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "ESPN scoreboard returned HTTP 400") ||
		strings.Contains(message, "ESPN scoreboard returned HTTP 403") ||
		strings.Contains(message, "ESPN scoreboard returned HTTP 429") ||
		strings.Contains(message, "ESPN scoreboard returned HTTP 5") ||
		strings.Contains(message, "ESPN scoreboard returned no week")
}

func summarizeResponseBody(body []byte) string {
	message := strings.TrimSpace(string(body))
	if message == "" {
		return "empty response"
	}
	if len(message) > 240 {
		return message[:240] + "..."
	}
	return message
}

func fetchAPTop25Week(client *http.Client, apiBase, pollID, week, pageURL string) ([]apTop25Rank, string, bool, error) {
	endpoint, err := url.Parse(strings.TrimRight(apiBase, "/") + "/top25PollResult")
	if err != nil {
		return nil, "", false, err
	}
	query := endpoint.Query()
	query.Set("top25PollId", pollID)
	query.Set("week", week)
	endpoint.RawQuery = query.Encode()
	apiRequest, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, "", false, err
	}
	apiRequest.Header.Set("Content-Type", "application/json")
	apiRequest.Header.Set("Origin", "https://apnews.com")
	apiRequest.Header.Set("Referer", pageURL)
	apiResponse, err := client.Do(apiRequest)
	if err != nil {
		return nil, "", false, err
	}
	defer apiResponse.Body.Close()
	if apiResponse.StatusCode == http.StatusNoContent || apiResponse.StatusCode == http.StatusBadRequest || apiResponse.StatusCode == http.StatusNotFound {
		return nil, "", false, nil
	}
	if apiResponse.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("AP rankings API returned HTTP %d", apiResponse.StatusCode)
	}

	var result apTop25Result
	if err := json.NewDecoder(apiResponse.Body).Decode(&result); err != nil {
		return nil, "", false, err
	}
	if len(result.Ranks) == 0 {
		return nil, "", false, nil
	}
	sort.Slice(result.Ranks, func(i, j int) bool { return result.Ranks[i].Rank < result.Ranks[j].Rank })
	if result.Week == "" {
		result.Week = week
	}
	return result.Ranks, result.Week, true, nil
}

func firstMatch(pattern *regexp.Regexp, data []byte) string {
	matches := pattern.FindSubmatch(data)
	if len(matches) < 2 {
		return ""
	}
	return string(matches[1])
}
