package parse

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/data"
	"aecheck-data-process/internal/types"
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// GetRecentLinks returns the character(buddy) links in specified date range
func GetRecentLinks(startDate time.Time, endDate time.Time, parseBuddy bool) ([]string, error) {
	url := "https://anothereden.wiki/w/Characters"

	resp, err := common.GetDataFromURL(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return nil, err
	}

	characterLinks := []string{}
	doc.Find("tr.character-row-entry").Each(func(i int, s *goquery.Selection) {
		// data-type이 존재하면 캐릭터, 아니면 버디
		data, isCharacter := s.Attr("data-type")
		if data == "" {
			isCharacter = false
		}
		isValid := (isCharacter != parseBuddy)
		if isValid {
			// 4번째 열의 날짜를 확인해서 n일 이내면 추가한다.
			date := s.Find("td").Eq(3).Text()
			dateTime, err := time.Parse("2006-01-02", date)
			if err != nil {
				return
			}

			if dateTime.After(startDate) && dateTime.Before(endDate) {
				if href, exists := s.Find("a").Attr("href"); exists {
					characterLinks = append(characterLinks, "https://anothereden.wiki"+href)
				}
			}
		}
	})

	return characterLinks, nil
}

func getRedirectURL(doc *goquery.Document, wikiURL string) string {
	redirectURL := ""
	if link, exists := doc.Find("link[rel='canonical']").Attr("href"); exists {
		redirectURL = link
	} else {
		redirectURL = wikiURL
	}
	return redirectURL
}

// ExtractCharacterInfoFromAEWikiURL extracts English name and style from character URL
func ExtractCharacterInfoFromAEWikiURL(wikiURL string) (*types.CharacterInfoFromAEWikiURL, error) {
	doc, err := data.GetDocumentFromURL(wikiURL)
	if err != nil {
		return nil, common.WrapErrorWithContext("ExtractCharacterInfoFromAEWikiURL", err)
	}

	redirectURL := getRedirectURL(doc, wikiURL)

	title := strings.TrimPrefix(redirectURL, "https://anothereden.wiki/w/")
	info := &types.CharacterInfoFromAEWikiURL{}

	// 이름 추출 (괄호 이전까지)
	if idx := strings.Index(title, "("); idx != -1 {
		info.EnglishName = strings.TrimSpace(strings.ReplaceAll(title[:idx], "_", " "))
	} else {
		info.EnglishName = strings.TrimSpace(strings.ReplaceAll(title, "_", " "))
	}

	// 이시층 여부 확인
	isAlter := strings.Contains(title, constants.AEWIKI_ALTER_SUFFIX)
	info.IsAlter = isAlter
	if isAlter {
		info.EnglishName += " (Alter)"
	}

	// 스타일 확인
	switch {
	case strings.Contains(title, "(Another_Style)"):
		info.Style = types.StyleAS
	case strings.Contains(title, "(Extra_Style)"):
		info.Style = types.StyleES
	default:
		info.Style = types.StyleNS
	}

	return info, nil
}

func GetCharacterInfo(wikiURL string) (*types.CharacterInfoFromAEWiki, error) {
	baseInfo, err := ExtractCharacterInfoFromAEWikiURL(wikiURL)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetCharacterInfo", err)
	}

	doc, err := data.GetDocumentFromURL(wikiURL)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetCharacterInfo", err)
	}

	info := &types.CharacterInfoFromAEWiki{
		CharacterInfoFromAEWikiURL: *baseInfo,
		WikiURL:                    getRedirectURL(doc, wikiURL),
	}

	generalData := doc.Find("article[title='General Data'] td")

	info.IsAwaken = logic.CheckAwaken(generalData.Eq(0).Text())
	info.LightShadow = logic.ClassifyLightShadow(generalData.Eq(5).Text())
	info.Category = logic.ClassifyCategory(generalData.Eq(6).Text())

	element := strings.TrimSpace(generalData.Eq(1).Text())
	var personalities []string
	generalData.Eq(7).Find("a").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		text = strings.Map(func(r rune) rune {
			if r == '-' || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
				return r
			}
			return -1
		}, text)
		personalities = append(personalities, text)
	})
	if strings.Contains(element, "None") {
		personalities = append(personalities, "None")
	}
	info.Personalities = personalities

	otherData := doc.Find("article[title='Other Data'] td")
	code := 0
	fmt.Sscanf(strings.TrimSpace(otherData.Eq(1).Text()), "%d", &code)
	info.GameID = code
	dateIndex := 6
	if info.IsAlter || constants.HIDDEN_NAMES[info.EnglishName] != "" {
		dateIndex = 9
	}

	updateDateStr, err := logic.ExtractDate(strings.TrimSpace(otherData.Eq(dateIndex).Text()))
	if err != nil {
		return nil, err
	}
	info.UpdateDate = updateDateStr

	manifestTableRow := doc.Find("div.character-weapon table tr").Eq(1)
	// 이 안에 th 태그가 있고 colspan 속성이 있으면 maxManifest 0
	manifestTableRow.Find("th").Each(func(i int, s *goquery.Selection) {
		if colspan, exists := s.Attr("colspan"); exists {
			if colspan == "0" {
				info.MaxManifest = 0
			}
		}
	})
	// 아니면 2번째 td 태그를 읽고 거기서 _(Enemy)가 아닌 링크를 추출
	manifestWeaponLink := ""
	tdSelection := manifestTableRow.Find("td")
	if tdSelection.Length() > 1 {
		tdSelection.Eq(1).Find("a").Each(func(i int, s *goquery.Selection) {
			if !strings.Contains(s.Text(), "(Enemy)") {
				manifestWeaponLink = "https://anothereden.wiki" + s.AttrOr("href", "")
			}
		})
		if manifestWeaponLink == "" {
			return nil, common.WrapErrorWithContext("no manifest weapon link found", nil)
		}
		info.IsManifestCustom = checkCustomManifest(manifestWeaponLink)
		if strings.Contains(tdSelection.Eq(5).Text(), "True Manifest") {
			info.MaxManifest = 2
		} else {
			info.MaxManifest = 1
		}
	} else {
		info.MaxManifest = 0
		info.IsManifestCustom = false
	}

	characterClassTable := doc.Find("div.character-class td")

	bookEndpoint := characterClassTable.Eq(7).Find("a").Eq(0).AttrOr("href", "")
	bookLink := "https://anothereden.wiki" + bookEndpoint

	className := strings.Split(strings.TrimSpace(characterClassTable.Eq(7).Text()), " ...▽ ")[0]
	// Remove newlines and control characters
	className = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, className)
	// Remove (free), (paid) etc.
	if idx := strings.Index(className, "("); idx != -1 {
		className = strings.TrimSpace(className[:idx])
	}
	info.EnglishClassName = className
	info.Dungeons, err = getDungeonsFromAEWiki(info.Style, info.IsAlter, bookLink)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func checkCustomManifest(manifestWeaponLink string) bool {
	resp, err := common.GetDataFromURL(manifestWeaponLink)
	if err != nil {
		return false
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return false
	}

	return strings.Contains(doc.Text(), "Weapon Tempering")
}

func getDungeonsFromAEWiki(style types.AEStyle, IsAlter bool, bookLink string) ([]string, error) {
	switch style {
	case types.StyleAS:
		return []string{"Treatise"}, nil
	case types.StyleES:
		return []string{"Codex"}, nil
	case types.StyleFOUR:
		return []string{}, nil
	}

	if IsAlter {
		return []string{"Opus"}, nil
	}

	doc, err := data.GetDocumentFromURL(bookLink)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var dungeons []string
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if !strings.HasPrefix(text, "Obtained from") && strings.Contains(text, "(VH)") {
			dungeons = append(dungeons, strings.Split(text, " (")[0])
		}
	})

	if len(dungeons) == 0 {
		return []string{"In-game"}, nil
	}
	return dungeons, nil
}

func GetBuddyInfoFromAEWiki(wikiURL string) (*types.BuddyInfoFromAEWiki, error) {
	doc, err := data.GetDocumentFromURL(wikiURL)
	if err != nil {
		return nil, common.WrapErrorWithContext("GetBuddyInfoFromAEWiki", err)
	}

	// 2. Redirect URL에서 이름, 이시층, 스타일을 찾는다.
	title := strings.TrimPrefix(wikiURL, constants.AEWIKI_BASE_URL)

	info := &types.BuddyInfoFromAEWiki{
		WikiURL: wikiURL,
	}

	// 이름 추출 (괄호 이전까지)
	if idx := strings.Index(title, "("); idx != -1 {
		info.EnglishName = strings.TrimSpace(strings.ReplaceAll(title[:idx], "_", " "))
	} else {
		info.EnglishName = strings.TrimSpace(strings.ReplaceAll(title, "_", " "))
	}

	// URL decode
	info.EnglishName, err = url.QueryUnescape(info.EnglishName)
	if err != nil {
		return nil, err
	}

	// 스타일 확인
	if strings.Contains(title, "(Another_Style)") {
		info.Style = types.StyleAS
	} else if strings.Contains(title, "(Extra_Style)") {
		info.Style = types.StyleES
	} else {
		info.Style = types.StyleNS
	}

	sidekickImage := doc.Find("div.sidekick-icon img")
	var code int
	fmt.Sscanf(strings.Split(sidekickImage.AttrOr("alt", ""), " ")[0], "%d", &code)
	info.GameID = code

	sidekickOwnerLink := doc.Find("div.sidekick-owner a")
	if sidekickOwnerLink.Length() > 0 {
		partnerURL := "https://anothereden.wiki" + sidekickOwnerLink.AttrOr("href", "")
		partnerDoc, err := data.GetDocumentFromURL(partnerURL)
		if err != nil {
			return nil, common.WrapErrorWithContext("GetBuddyInfoFromAEWiki", err)
		}
		info.PartnerLink = getRedirectURL(partnerDoc, partnerURL)
	} else {
		info.PartnerLink = "None"
	}

	return info, nil
}
