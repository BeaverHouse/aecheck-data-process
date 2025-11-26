package logic

import (
	"fmt"
	"strings"
	"time"
)

func ExtractDate(rawString string) (string, error) {
	parsedDateString := strings.Split(rawString, " / ")[1]

	dateFormats := []string{
		"Jan 2, 2006",
		"January 2, 2006",
		"2 Jan, 2006",
		"Jan 2 2006",
		"2006-01-02",
	}

	var updateDate string
	for _, format := range dateFormats {
		t, err := time.Parse(format, parsedDateString)
		if err == nil {
			updateDate = t.Format("2006-01-02")
			break
		}
	}
	if updateDate == "" {
		return "", fmt.Errorf("cannot parse update date: %s", parsedDateString)
	}

	return updateDate, nil
}
