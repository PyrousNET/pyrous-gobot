package jokes

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type JokeFeedback struct {
	Timestamp time.Time
	UserID    string
	JokeID    string
	Rating    int
	Comment   string
	Joke      Joke
}

var feedbackFileMu sync.Mutex

var (
	legacyFeedbackHeader = []string{"timestamp_utc", "rating", "comment", "setup", "punchline", "provider", "model"}
	feedbackHeader       = []string{"timestamp_utc", "user_id", "rating", "comment", "setup", "punchline", "provider", "model", "joke_id"}
)

func AppendFeedback(path string, feedback JokeFeedback) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("joke feedback CSV path is empty")
	}
	if feedback.Rating < 1 || feedback.Rating > 5 {
		return fmt.Errorf("joke feedback rating %d is outside 1-5", feedback.Rating)
	}
	if feedback.Timestamp.IsZero() {
		feedback.Timestamp = time.Now().UTC()
	} else {
		feedback.Timestamp = feedback.Timestamp.UTC()
	}

	feedbackFileMu.Lock()
	defer feedbackFileMu.Unlock()

	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read joke feedback CSV: %w", err)
	}
	if len(data) > 0 {
		records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
		if err != nil {
			return fmt.Errorf("read existing joke feedback CSV: %w", err)
		}
		if len(records) == 0 {
			return fmt.Errorf("existing joke feedback CSV has no header")
		}
		switch {
		case sameCSVRecord(records[0], legacyFeedbackHeader):
			if err := upgradeFeedbackCSV(path, records); err != nil {
				return err
			}
		case !sameCSVRecord(records[0], feedbackHeader):
			return fmt.Errorf("existing joke feedback CSV has an unsupported header")
		}
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open joke feedback CSV: %w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat joke feedback CSV: %w", err)
	}
	writer := csv.NewWriter(file)
	if info.Size() == 0 {
		if err := writer.Write(feedbackHeader); err != nil {
			return fmt.Errorf("write joke feedback CSV header: %w", err)
		}
	}
	if err := writer.Write([]string{
		feedback.Timestamp.Format(time.RFC3339Nano),
		feedback.UserID,
		fmt.Sprint(feedback.Rating),
		feedback.Comment,
		feedback.Joke.Setup,
		feedback.Joke.Punchline,
		feedback.Joke.Provider,
		feedback.Joke.Model,
		feedback.JokeID,
	}); err != nil {
		return fmt.Errorf("write joke feedback CSV row: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush joke feedback CSV: %w", err)
	}
	return nil
}

func sameCSVRecord(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

func upgradeFeedbackCSV(path string, records [][]string) error {
	temp, err := os.CreateTemp(filepath.Dir(path), ".joke-feedback-*.tmp")
	if err != nil {
		return fmt.Errorf("create upgraded joke feedback CSV: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("set upgraded joke feedback CSV permissions: %w", err)
	}

	writer := csv.NewWriter(temp)
	if err := writer.Write(feedbackHeader); err != nil {
		temp.Close()
		return fmt.Errorf("write upgraded joke feedback CSV header: %w", err)
	}
	for _, record := range records[1:] {
		if len(record) != len(legacyFeedbackHeader) {
			temp.Close()
			return fmt.Errorf("existing joke feedback CSV row has %d fields; want %d", len(record), len(legacyFeedbackHeader))
		}
		upgraded := []string{record[0], "", record[1], record[2], record[3], record[4], record[5], record[6], ""}
		if err := writer.Write(upgraded); err != nil {
			temp.Close()
			return fmt.Errorf("write upgraded joke feedback CSV row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		temp.Close()
		return fmt.Errorf("flush upgraded joke feedback CSV: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync upgraded joke feedback CSV: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close upgraded joke feedback CSV: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace joke feedback CSV with upgraded format: %w", err)
	}
	return nil
}
