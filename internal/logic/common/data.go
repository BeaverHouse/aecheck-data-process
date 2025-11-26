package common

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Gets data from a URL.
//
// If the URL is invalid or the data is empty, it returns an error.
func GetDataFromURL(url string) ([]byte, error) {
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, WrapErrorWithContext("GetDataFromURL", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, WrapErrorWithContext("GetDataFromURL", fmt.Errorf("invalid status code: %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, WrapErrorWithContext("GetDataFromURL", err)
	} else if len(body) == 0 {
		return nil, WrapErrorWithContext("GetDataFromURL", fmt.Errorf("the data from URL is empty: %s", url))
	}
	LogInfo("Response successfully fetched", zap.String("url", url), zap.Duration("duration", time.Since(start)))
	return body, nil
}
