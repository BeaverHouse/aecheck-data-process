package parse

import (
	"aecheck-data-process/internal/logic/common"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetHTMLcontent(announceID int, language string) (string, error) {
	url := fmt.Sprintf("https://api-ap.another-eden.games/asset/notice_v2/view/%d?language=%s", announceID, language)

	resp, err := common.GetDataFromURL(url)
	if err != nil {
		return "", err
	}
	// Parse HTML content using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp)))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Extract text inside li tags within body
	var texts []string
	doc.Find("body li").Each(func(i int, s *goquery.Selection) {
		texts = append(texts, strings.TrimSpace(s.Text()))
	})
	text := strings.Join(texts, "\n")

	text = strings.TrimSpace(text)

	return text, nil
}
