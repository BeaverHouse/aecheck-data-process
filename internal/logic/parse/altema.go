package parse

import (
	"aecheck-data-process/internal/logic/common"
	"bytes"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func GetAltemaLink(characterName string, isTales bool) (string, error) {
	// 이름에 공백이 있으면 첫 단어를 잘라서 검색어로 사용, 아니면 그대로 검색어로 사용
	searchName := ""
	if strings.Contains(characterName, " ") {
		searchName = strings.Split(characterName, " ")[0]
	} else {
		searchName = characterName
	}

	resp, err := common.GetDataFromURL("https://altema.jp/anaden/charalist-2")
	if err != nil {
		return "", err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return "", err
	}

	var link string
	doc.Find("tr.charaTr").Each(func(i int, s *goquery.Selection) {
		obtain := strings.TrimSpace(s.Find("td:nth-child(5)").Text())
		if isTales && obtain != "協奏" {
			return
		} else if !isTales && obtain == "協奏" {
			return
		}
		// 첫 열의 텍스트에 searchName이 있으면 거기 있는 링크를 반환
		if strings.TrimSpace(strings.ReplaceAll(s.Find("td:nth-child(1)").Text(), "【New!】", "")) == searchName {
			link, _ = s.Find("td:nth-child(1) a").Attr("href")
		}
	})

	if link == "" {
		return "", common.WrapErrorWithContext("no link found", nil)
	}

	return "https://altema.jp" + link, nil
}
