package jokes

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestAppendFeedbackWritesHeaderOnceAndPreservesFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "joke-feedback.csv")
	feedback := JokeFeedback{
		Timestamp: time.Date(2026, time.November, 9, 17, 0, 0, 0, time.UTC),
		UserID:    "user-1",
		JokeID:    "joke-123",
		Rating:    2,
		Comment:   "The punchline doesn't follow, and has, commas",
		Joke:      Joke{Setup: "A setup", Punchline: "A punchline", Provider: "openai joke generator", Model: "gpt-4.1-mini"},
	}
	if err := AppendFeedback(path, feedback); err != nil {
		t.Fatalf("AppendFeedback() error = %v", err)
	}
	feedback.Rating = 5
	feedback.Comment = "Much better"
	if err := AppendFeedback(path, feedback); err != nil {
		t.Fatalf("second AppendFeedback() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"timestamp_utc", "user_id", "rating", "comment", "setup", "punchline", "provider", "model", "joke_id"},
		{"2026-11-09T17:00:00Z", "user-1", "2", "The punchline doesn't follow, and has, commas", "A setup", "A punchline", "openai joke generator", "gpt-4.1-mini", "joke-123"},
		{"2026-11-09T17:00:00Z", "user-1", "5", "Much better", "A setup", "A punchline", "openai joke generator", "gpt-4.1-mini", "joke-123"},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("CSV records = %#v, want %#v", records, want)
	}
}

func TestAppendFeedbackUpgradesLegacyCSVBeforeAppending(t *testing.T) {
	path := filepath.Join(t.TempDir(), "joke-feedback.csv")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	legacy := csv.NewWriter(file)
	for _, record := range [][]string{
		{"timestamp_utc", "rating", "comment", "setup", "punchline", "provider", "model"},
		{"2026-11-08T17:00:00Z", "3", "Old feedback", "Old setup", "Old punchline", "openai joke generator", "gpt-4.1-mini"},
	} {
		if err := legacy.Write(record); err != nil {
			t.Fatal(err)
		}
	}
	legacy.Flush()
	if err := legacy.Error(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	newFeedback := JokeFeedback{
		Timestamp: time.Date(2026, time.November, 9, 17, 0, 0, 0, time.UTC),
		UserID:    "user-2",
		JokeID:    "joke-456",
		Rating:    5,
		Comment:   "New feedback",
		Joke:      Joke{Setup: "New setup", Punchline: "New punchline", Provider: "reddit"},
	}
	if err := AppendFeedback(path, newFeedback); err != nil {
		t.Fatalf("AppendFeedback() with legacy CSV error = %v", err)
	}

	upgraded, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer upgraded.Close()
	records, err := csv.NewReader(upgraded).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"timestamp_utc", "user_id", "rating", "comment", "setup", "punchline", "provider", "model", "joke_id"},
		{"2026-11-08T17:00:00Z", "", "3", "Old feedback", "Old setup", "Old punchline", "openai joke generator", "gpt-4.1-mini", ""},
		{"2026-11-09T17:00:00Z", "user-2", "5", "New feedback", "New setup", "New punchline", "reddit", "", "joke-456"},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("upgraded CSV records = %#v, want %#v", records, want)
	}
}

func TestJokeFeedbackProfileTurnsRatingsIntoCachedGenerationGuidance(t *testing.T) {
	var profile JokeFeedbackProfile
	profile.Add(1, "The pun felt forced and the punchline made no sense")
	profile.Add(3, "It was funny but not laugh out loud funny")
	profile.Add(5, "The clever twist made this one work")

	guidance := profile.GenerationGuidance()
	for _, want := range []string{"3 rating(s)", "clear logical connection", "sharper twist", "use puns only when they feel natural", "surprising twists"} {
		if !strings.Contains(guidance, want) {
			t.Errorf("generation guidance %q does not include %q", guidance, want)
		}
	}
}

func TestLoadJokeFeedbackProfileReadsLegacyCSV(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy-feedback.csv")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := csv.NewWriter(file)
	for _, row := range [][]string{
		{"timestamp_utc", "rating", "comment", "setup", "punchline", "provider", "model"},
		{"2026-11-08T17:00:00Z", "2", "The punchline was confusing", "Setup", "Punchline", "reddit", ""},
	} {
		if err := writer.Write(row); err != nil {
			t.Fatal(err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	profile, err := LoadJokeFeedbackProfile(path)
	if err != nil {
		t.Fatalf("LoadJokeFeedbackProfile() error = %v", err)
	}
	if profile.Count != 1 || profile.RatingTotal != 2 || profile.ConfusingLow != 1 {
		t.Fatalf("profile = %#v, want one low confusing rating", profile)
	}
}
