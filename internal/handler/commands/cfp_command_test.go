package commands

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pyrousnet/pyrous-gobot/internal/cache"
	"github.com/pyrousnet/pyrous-gobot/internal/settings"
)

func TestCFPCommandIsRegistered(t *testing.T) {
	commands := NewCommands(settings.SetupMockSettings(sync.RWMutex{}, settings.CommandSettings{CommandTrigger: "!"}), nil, &cache.MockCache{}, nil)
	command, err := commands.NewBotCommand("!cfp", "test")
	if err != nil {
		t.Fatalf("NewBotCommand(!cfp) error = %v", err)
	}
	if command.target != "cfp" {
		t.Fatalf("command target = %q, want cfp", command.target)
	}
	doc := (BotCommandHelp{}).Cfp(BotCommand{})
	if !strings.Contains(doc.Help, "!cfp") {
		t.Fatalf("CFP help = %q, want !cfp usage", doc.Help)
	}
}

func TestCFPCacheTTLRetriesDuringCommitteeReleaseWindow(t *testing.T) {
	location, err := time.LoadLocation("America/Denver")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		now  time.Time
		want time.Duration
	}{
		{
			name: "before Tuesday release",
			now:  time.Date(2026, time.November, 10, 18, 59, 0, 0, location),
			want: time.Minute,
		},
		{
			name: "immediately after Tuesday release",
			now:  time.Date(2026, time.November, 10, 19, 0, 0, 0, location),
			want: cfpPostReleaseCacheTTL,
		},
		{
			name: "during Wednesday retry window",
			now:  time.Date(2026, time.November, 11, 6, 59, 0, 0, location),
			want: cfpPostReleaseCacheTTL,
		},
		{
			name: "after Wednesday retry window",
			now:  time.Date(2026, time.November, 11, 7, 1, 0, 0, location),
			want: durationUntilNextCFPRankings(time.Date(2026, time.November, 11, 7, 1, 0, 0, location)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfpCacheTTL(tt.now); got != tt.want {
				t.Fatalf("cfpCacheTTL(%s) = %s, want %s", tt.now, got, tt.want)
			}
		})
	}
}

func TestFetchCachedCFPRankingsIgnoresPreviousCacheVersion(t *testing.T) {
	c := cache.GetLocalCache()
	oldValue, err := json.Marshal(cfpRankingCache{
		Week:  "Week 10",
		Ranks: []cfpRank{{Rank: 1, TeamID: 1, TeamName: "Old ranking"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	c.Put("cfp-rankings:v2:2026", string(oldValue))

	fetched := false
	ranks, week, err := fetchCachedCFPRankingsWithFetcher(nil, c, "test-key", 2026, func(*http.Client, string, int) ([]cfpRank, string, error) {
		fetched = true
		return []cfpRank{{Rank: 1, TeamID: 2, TeamName: "Updated ranking"}}, "Week 11", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !fetched {
		t.Fatal("fetcher was not called; prior cached rankings were reused")
	}
	if week != "Week 11" || len(ranks) != 1 || ranks[0].TeamName != "Updated ranking" {
		t.Fatalf("got %q rankings %#v, want updated Week 11 rankings", week, ranks)
	}
}
