package external

import (
	"bytes"

	"aecheck-data-process/internal/logic/common"

	"github.com/BeaverHouse/go-common/errorhandle"
	"github.com/PuerkitoBio/goquery"
)

// GetDocumentFromURL fetches HTML content from a URL and returns a goquery document
func GetDocumentFromURL(url string) (*goquery.Document, error) {
	resp, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, errorhandle.Wrap(errorhandle.KindUnavailable, "WIKI_FETCH_FAILED", "GetDocumentFromURL", err)
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return nil, errorhandle.Wrap(errorhandle.KindInternal, "WIKI_DOCUMENT_PARSE_FAILED", "GetDocumentFromURL", err)
	}
	return doc, nil
}
