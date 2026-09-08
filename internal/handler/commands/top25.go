package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"github.com/pyrousnet/pyrous-gobot/internal/comms"
	"github.com/pyrousnet/pyrous-gobot/internal/users"
)

const apTop25URL = "https://apnews.com/hub/ap-top-25-college-football-poll"

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

	teams, week, err := fetchAPTop25(http.DefaultClient)
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
		return fmt.Sprintf("▲ %d", trend)
	case trend < 0:
		return fmt.Sprintf("▼ %d", -trend)
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
