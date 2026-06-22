package logic

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

var dateFormats = []string{
	"Jan 2, 2006",
	"January 2, 2006",
	"2 Jan, 2006",
	"Jan 2 2006",
	"2006-01-02",
}

var datePatterns = []*regexp.Regexp{
	regexp.MustCompile(`[A-Z][a-z]+ \d{1,2}, \d{4}`),
	regexp.MustCompile(`\d{1,2} [A-Z][a-z]+, \d{4}`),
	regexp.MustCompile(`[A-Z][a-z]+ \d{1,2} \d{4}`),
	regexp.MustCompile(`\d{4}-\d{2}-\d{2}`),
}

func ExtractDate(rawString string) (string, error) {
	candidates := extractDateCandidates(rawString)
	if len(candidates) == 0 {
		return "", fmt.Errorf("cannot parse update date: empty input")
	}

	for _, candidate := range candidates {
		for _, format := range dateFormats {
			t, err := time.Parse(format, candidate)
			if err == nil {
				return t.Format("2006-01-02"), nil
			}
		}
	}

	return "", fmt.Errorf("cannot parse update date: %s", strings.TrimSpace(rawString))
}

func extractDateCandidates(rawString string) []string {
	text := strings.TrimSpace(rawString)
	if text == "" {
		return nil
	}

	candidates := []string{text}
	if parts := strings.Split(text, " / "); len(parts) > 1 {
		candidates = append(candidates, strings.TrimSpace(parts[len(parts)-1]))
	}

	for _, pattern := range datePatterns {
		candidates = append(candidates, pattern.FindAllString(text, -1)...)
	}

	return candidates
}
