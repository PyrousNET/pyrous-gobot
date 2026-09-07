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
			io.WriteString(w, `{"week":"Week 2","ranks":[{"rank":2,"teamName":"Beta"},{"rank":1,"teamName":"Alpha"}]}`)
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
	if teams[0].Rank != 1 || teams[0].TeamName != "Alpha" {
		t.Fatalf("teams were not sorted by rank: %#v", teams)
	}
}

func TestFormatAPTop25(t *testing.T) {
	message := formatAPTop25([]apTop25Rank{{Rank: 1, Trend: 2, TeamName: "Alpha"}, {Rank: 25, Trend: -3, TeamName: "Beta"}, {Rank: 12, TeamName: "Gamma"}}, "Week 2")
	want := []string{
		"### AP Top 25 (Week 2)",
		"| Rank | Team | Movement |",
		"| ---: | --- | :---: |",
		"| 1 | Alpha | ▲ 2 |",
		"| 25 | Beta | ▼ 3 |",
		"| 12 | Gamma | — |",
		"Source: [AP News — AP Top 25 Poll](" + apTop25URL + ")",
	}
	for _, line := range want {
		if !strings.Contains(message, line) {
			t.Errorf("formatted message missing %q:\n%s", line, message)
		}
	}
}
