package external

import (
	"bytes"
	"fmt"

	"aecheck-data-process/internal/logic/common"

	"github.com/PuerkitoBio/goquery"
)

// GetDocumentFromURL fetches HTML content from a URL and returns a goquery document
func GetDocumentFromURL(url string) (*goquery.Document, error) {
	resp, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("GetDocumentFromURL: %w", err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return nil, fmt.Errorf("GetDocumentFromURL: %w", err)
	}
	return doc, nil
}
