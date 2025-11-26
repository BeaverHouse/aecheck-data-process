package logic

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/types"
	"fmt"
	"net/http"
	"strings"

	"io"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func checkPageExists(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func FindAlterLink(info types.CharacterInfoFromAEWiki) string {
	// 1. style이 NS가 아닐 경우 "None"
	if info.Style != types.StyleNS {
		return "None"
	}

	// 2. style이 NS인데 이름에 isAlter가 true라면 링크에서 _(Alter)를 뺀 게 답이다.
	if info.IsAlter {
		return strings.Replace(info.WikiURL, constants.AEWIKI_ALTER_SUFFIX, "", 1)
	}

	// 3. style이 NS인데 이름에 isAlter가 false라면 링크 끝에 _(Alter)를 붙이고, 그 페이지가 실제로 존재하는 페이지라면 그게 답이다. 아니면 "None"이다.
	alterURL := info.WikiURL + constants.AEWIKI_ALTER_SUFFIX
	if checkPageExists(alterURL) {
		return alterURL
	}
	return "None"
}

func FindSeesaaLink(info types.CharacterInfoFromAEWiki, japaneseName string) string {
	seesaaURL := constants.SEESAA_BASE_URL + japaneseName
	switch info.Style {
	case types.StyleAS:
		seesaaURL += "%28アナザースタイル%29"
	case types.StyleES:
		seesaaURL += "%28エクストラスタイル%29"
	}
	// Encode URL to EUC-JP for Japanese characters
	seesaaURL = strings.ReplaceAll(seesaaURL, " ", "%20")

	// Convert Japanese characters to EUC-JP encoding
	reader := transform.NewReader(strings.NewReader(seesaaURL), japanese.EUCJP.NewEncoder())
	eucjpBytes, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	// Convert bytes to URL-encoded string
	var encodedParts []string
	for _, b := range eucjpBytes {
		if b >= 0x20 && b <= 0x7E {
			encodedParts = append(encodedParts, string(b))
		} else {
			encodedParts = append(encodedParts, fmt.Sprintf("%%%02X", b))
		}
	}
	seesaaURL = strings.Join(encodedParts, "")
	return seesaaURL
}
