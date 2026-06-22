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

	// 2. style이 NS인데 isAlter가 true라면 이 캐릭터가 Alter이고, 원본 URL을 반환해야 한다.
	if info.IsAlter {
		// 가명 URL인 경우(_(Alter)가 없음): 진명 기반 URL을 직접 구성
		// e.g. Nekoko → Necoco_(Alter) → 원본은 Necoco
		if !strings.Contains(info.WikiURL, constants.AEWIKI_ALTER_SUFFIX) {
			trueName := strings.TrimSuffix(info.EnglishName, " (Alter)")
			trueName = strings.ReplaceAll(trueName, " ", "_")
			return constants.AEWIKI_BASE_URL + trueName
		}
		return strings.Replace(info.WikiURL, constants.AEWIKI_ALTER_SUFFIX, "", 1)
	}

	// 3. style이 NS인데 isAlter가 false라면 링크 끝에 _(Alter)를 붙이고, 그 페이지가 실제로 존재하는 페이지라면 그게 답이다. 아니면 "None"이다.
	alterURL := info.WikiURL + constants.AEWIKI_ALTER_SUFFIX
	if checkPageExists(alterURL) {
		return alterURL
	}
	return "None"
}

func encodeSeesaaURL(raw string) string {
	raw = strings.ReplaceAll(raw, " ", "%20")
	reader := transform.NewReader(strings.NewReader(raw), japanese.EUCJP.NewEncoder())
	eucjpBytes, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	var parts []string
	for _, b := range eucjpBytes {
		if b >= 0x20 && b <= 0x7E {
			parts = append(parts, string(b))
		} else {
			parts = append(parts, fmt.Sprintf("%%%02X", b))
		}
	}
	return strings.Join(parts, "")
}

func FindSeesaaLink(info types.CharacterInfoFromAEWiki, japaneseName string) string {
	seesaaURL := constants.SEESAA_BASE_URL + japaneseName
	switch info.Style {
	case types.StyleAS:
		seesaaURL += "%28アナザースタイル%29"
	case types.StyleES:
		seesaaURL += "%28エクストラスタイル%29"
	}
	return encodeSeesaaURL(seesaaURL)
}

func FindBuddySeesaaLink(style types.AEStyle, japaneseName string) string {
	seesaaURL := constants.SEESAA_BASE_URL + japaneseName
	switch style {
	case types.StyleAS:
		seesaaURL += "%28アナザースタイル%29"
	case types.StyleES:
		seesaaURL += "%28エクストラスタイル%29"
	}
	return encodeSeesaaURL(seesaaURL)
}
