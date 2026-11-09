package jokes

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestAppendFeedbackWritesHeaderOnceAndPreservesFields(t *testing.T) {
	path := filepath.Join(t.TempDir(), "joke-feedback.csv")
	feedback := JokeFeedback{
		Timestamp: time.Date(2026, time.November, 9, 17, 0, 0, 0, time.UTC),
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
		{"timestamp_utc", "rating", "comment", "setup", "punchline", "provider", "model"},
		{"2026-11-09T17:00:00Z", "2", "The punchline doesn't follow, and has, commas", "A setup", "A punchline", "openai joke generator", "gpt-4.1-mini"},
		{"2026-11-09T17:00:00Z", "5", "Much better", "A setup", "A punchline", "openai joke generator", "gpt-4.1-mini"},
	}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("CSV records = %#v, want %#v", records, want)
	}
}
