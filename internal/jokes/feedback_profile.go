package jokes

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const JokeFeedbackProfileCacheKey = "joke-feedback:profile"

type JokeFeedbackProfile struct {
	Count           int `json:"count"`
	RatingTotal     int `json:"rating_total"`
	ConfusingLow    int `json:"confusing_low"`
	Underwhelming   int `json:"underwhelming"`
	ForcedWordplay  int `json:"forced_wordplay"`
	StaleLow        int `json:"stale_low"`
	VerboseLow      int `json:"verbose_low"`
	WordplayLiked   int `json:"wordplay_liked"`
	UnexpectedLiked int `json:"unexpected_liked"`
}

func (profile *JokeFeedbackProfile) Add(rating int, comment string) {
	if profile == nil || rating < 1 || rating > 5 {
		return
	}
	profile.Count++
	profile.RatingTotal += rating
	comment = strings.ToLower(comment)
	if containsAny(comment, "not laugh", "not funny", "not hilarious", "underwhelming", "meh", "funny but") {
		profile.Underwhelming++
	}
	if rating <= 2 {
		if containsAny(comment, "sense", "follow", "connect", "confus", "random", "unrelated") {
			profile.ConfusingLow++
		}
		if containsAny(comment, "pun", "wordplay", "corny", "cheesy", "forced") {
			profile.ForcedWordplay++
		}
		if containsAny(comment, "stock", "clich", "predictable", "overused", "heard that") {
			profile.StaleLow++
		}
		if containsAny(comment, "long", "wordy", "too much", "too many words") {
			profile.VerboseLow++
		}
	}
	if rating >= 4 {
		if containsAny(comment, "pun", "wordplay", "clever") {
			profile.WordplayLiked++
		}
		if containsAny(comment, "surpris", "twist", "unexpected") {
			profile.UnexpectedLiked++
		}
	}
}

func (profile JokeFeedbackProfile) GenerationGuidance() string {
	if profile.Count == 0 {
		return ""
	}
	average := float64(profile.RatingTotal) / float64(profile.Count)
	guidance := []string{fmt.Sprintf("Community feedback: %d rating(s), averaging %.1f/5.", profile.Count, average)}
	if average < 3 {
		guidance = append(guidance, "Ratings trend low; prioritize a clear, satisfying punchline and avoid forced humor.")
	} else if average >= 4 {
		guidance = append(guidance, "Ratings are strong; preserve concise, original jokes while varying the comic idea.")
	}
	if profile.ConfusingLow > 0 {
		guidance = append(guidance, "Low ratings flagged confusing payoffs; ensure a clear logical connection between setup and punchline.")
	}
	if profile.Underwhelming > 0 {
		guidance = append(guidance, "Feedback found some jokes merely adequate; aim for a sharper twist or a more satisfying laugh.")
	}
	if profile.ForcedWordplay > 0 {
		guidance = append(guidance, "Low ratings flagged forced wordplay; use puns only when they feel natural.")
	}
	if profile.StaleLow > 0 {
		guidance = append(guidance, "Low ratings flagged stale jokes; avoid clichés and familiar stock punchlines.")
	}
	if profile.VerboseLow > 0 {
		guidance = append(guidance, "Low ratings flagged wordiness; keep the setup and punchline concise.")
	}
	if profile.WordplayLiked > 0 {
		guidance = append(guidance, "High ratings appreciated clever wordplay when it fits naturally.")
	}
	if profile.UnexpectedLiked > 0 {
		guidance = append(guidance, "High ratings appreciated surprising twists with a clear payoff.")
	}
	return strings.Join(guidance, " ")
}

func LoadJokeFeedbackProfile(path string) (JokeFeedbackProfile, error) {
	var profile JokeFeedbackProfile
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return profile, nil
	}
	if err != nil {
		return profile, fmt.Errorf("open joke feedback CSV: %w", err)
	}
	defer file.Close()

	records, err := csv.NewReader(file).ReadAll()
	if err != nil {
		return profile, fmt.Errorf("read joke feedback CSV: %w", err)
	}
	if len(records) == 0 {
		return profile, nil
	}
	columns := make(map[string]int, len(records[0]))
	for i, name := range records[0] {
		columns[name] = i
	}
	ratingColumn, hasRating := columns["rating"]
	commentColumn, hasComment := columns["comment"]
	if !hasRating || !hasComment {
		return profile, errors.New("joke feedback CSV is missing rating or comment columns")
	}
	for _, record := range records[1:] {
		if ratingColumn >= len(record) || commentColumn >= len(record) {
			continue
		}
		rating, err := strconv.Atoi(record[ratingColumn])
		if err != nil {
			continue
		}
		profile.Add(rating, record[commentColumn])
	}
	return profile, nil
}

func containsAny(value string, terms ...string) bool {
	for _, term := range terms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}
