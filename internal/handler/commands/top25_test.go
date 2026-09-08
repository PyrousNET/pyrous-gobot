package commands

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFetchAPTop25(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/hub/ap-top-25-college-football-poll":
			io.WriteString(w, `<bsp-top-poll-results-module data-top25-poll-id="poll-123" data-week="Week 2" data-api-base-url="`+"http://"+r.Host+`"></bsp-top-poll-results-module>`)
		case "/top25PollResult":
			if r.URL.Query().Get("top25PollId") != "poll-123" || r.URL.Query().Get("week") != "Week 2" {
				t.Errorf("unexpected query: %s", r.URL.RawQuery)
			}
			io.WriteString(w, `{"week":"Week 2","ranks":[{"rank":2,"teamName":"Beta","win":8,"loss":4,"tie":0},{"rank":1,"teamName":"Alpha","win":10,"loss":2,"tie":1}]}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	teams, week, err := fetchAPTop25FromURL(server.Client(), server.URL+"/hub/ap-top-25-college-football-poll")
	if err != nil {
		t.Fatalf("fetchAPTop25FromURL() error = %v", err)
	}
	if week != "Week 2" || len(teams) != 2 {
		t.Fatalf("week = %q, teams = %#v", week, teams)
	}
	if teams[0].Rank != 1 || teams[0].TeamName != "Alpha" || formatAPRecord(teams[0]) != "10-2-1" {
		t.Fatalf("teams were not sorted by rank: %#v", teams)
	}
}

func TestFetchAPTop25FallsBackWhenCurrentPollIsUnavailable(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hub/ap-top-25-college-football-poll" {
			io.WriteString(w, `<bsp-top-poll-results-module data-top25-poll-id="poll-123" data-week="Week 2" data-api-base-url="`+server.URL+`"></bsp-top-poll-results-module><option value="Week 1" data-close-at="1">Week 1</option>`)
			return
		}
		if r.URL.Query().Get("week") == "Week 2" {
			http.Error(w, `{"message":"Result not available for this week"}`, http.StatusBadRequest)
			return
		}
		io.WriteString(w, `{"week":"Week 1","ranks":[{"rank":1,"teamName":"Alpha"}]}`)
	}))
	defer server.Close()

	teams, week, err := fetchAPTop25FromURL(server.Client(), server.URL+"/hub/ap-top-25-college-football-poll")
	if err != nil {
		t.Fatalf("fetchAPTop25FromURL() error = %v", err)
	}
	if week != "Week 1" || len(teams) != 1 || teams[0].TeamName != "Alpha" {
		t.Fatalf("fallback result = week %q, teams %#v", week, teams)
	}
}

func TestFormatAPTop25(t *testing.T) {
	message := formatAPTop25([]apTop25Rank{{Rank: 1, Trend: 2, TeamName: "Alpha", Wins: 10, Losses: 2, Ties: 1}, {Rank: 25, Trend: -3, TeamName: "Beta", Wins: 8, Losses: 4}, {Rank: 12, TeamName: "Gamma", Wins: 6, Losses: 6}}, "Week 2")
	want := []string{
		"### AP Top 25 (Week 2)",
		"| Rank | Team | Record | Movement |",
		"| ---: | --- | :---: | :---: |",
		"| 1 | Alpha | 10-2-1 | 🟢 ▲ 2 |",
		"| 25 | Beta | 8-4-0 | 🔴 ▼ 3 |",
		"| 12 | Gamma | 6-6-0 | — |",
		"Source: [AP News — AP Top 25 Poll](" + apTop25URL + ")",
	}
	for _, line := range want {
		if !strings.Contains(message, line) {
			t.Errorf("formatted message missing %q:\n%s", line, message)
		}
	}
}
