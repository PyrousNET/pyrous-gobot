package jokes

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type JokeFeedback struct {
	Timestamp time.Time
	Rating    int
	Comment   string
	Joke      Joke
}

var feedbackFileMu sync.Mutex

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
		if err := writer.Write([]string{"timestamp_utc", "rating", "comment", "setup", "punchline", "provider", "model"}); err != nil {
			return fmt.Errorf("write joke feedback CSV header: %w", err)
		}
	}
	if err := writer.Write([]string{
		feedback.Timestamp.Format(time.RFC3339Nano),
		fmt.Sprint(feedback.Rating),
		feedback.Comment,
		feedback.Joke.Setup,
		feedback.Joke.Punchline,
		feedback.Joke.Provider,
		feedback.Joke.Model,
	}); err != nil {
		return fmt.Errorf("write joke feedback CSV row: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush joke feedback CSV: %w", err)
	}
	return nil
}
