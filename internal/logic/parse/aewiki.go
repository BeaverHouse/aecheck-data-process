package parse

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	data "aecheck-data-process/internal/logic/external"
	"aecheck-data-process/internal/types"
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

type RecentUpdateKind string

const (
	RecentUpdateReleased RecentUpdateKind = "released"
	RecentUpdateSidekick RecentUpdateKind = "sidekick"
	RecentUpdateStellar  RecentUpdateKind = "stellar"
)

type RecentUpdateLink struct {
	Text string
	URL  string
}

type RecentUpdateGroup struct {
	Kind  RecentUpdateKind
	Label string
	Links []RecentUpdateLink
}

func GetRecentUpdateGroups() ([]RecentUpdateGroup, error) {
	resp, err := common.GetDataFromURL(constants.AEWIKI_BASE_URL)
	if err != nil {
		return nil, err
	}
	return ExtractRecentUpdateGroups(resp)
}

func ExtractRecentUpdateGroups(html []byte) ([]RecentUpdateGroup, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, err
	}

	groups := []RecentUpdateGroup{}
	doc.Find("#main-page-new-content .char-time-block").Each(func(i int, s *goquery.Selection) {
		label, kind := recentUpdateLabel(s)
		if kind == "" {
			return
		}
		links := recentUpdateLinks(s)
		if len(links) == 0 {
			return
		}
		groups = append(groups, RecentUpdateGroup{Kind: kind, Label: label, Links: links})
	})
	if len(groups) == 0 {
		return nil, fmt.Errorf("no recent update groups found on AE Wiki home; solve Cloudflare challenge and retry")
	}
	return groups, nil
}

func recentUpdateLabel(s *goquery.Selection) (string, RecentUpdateKind) {
	var label string
	var kind RecentUpdateKind

	s.Find("div").EachWithBreak(func(i int, div *goquery.Selection) bool {
		if div.Find("a[href]").Length() > 0 {
			return true
		}
		text := normalizeWikiText(div.Text())
		if text == "" {
			return true
		}
		if classified := classifyRecentUpdateLabel(text); classified != "" {
			label = text
			kind = classified
			return false
		}
		return true
	})

	return label, kind
}

func classifyRecentUpdateLabel(label string) RecentUpdateKind {
	lower := strings.ToLower(label)
	switch {
	case strings.Contains(lower, "released:"):
		return RecentUpdateReleased
	case strings.Contains(lower, "sidekick:"):
		return RecentUpdateSidekick
	case strings.Contains(lower, "stellar"):
		return RecentUpdateStellar
	default:
		return ""
	}
}

func recentUpdateLinks(s *goquery.Selection) []RecentUpdateLink {
	seen := map[string]bool{}
	links := []RecentUpdateLink{}
	s.Find("a[href]").Each(func(i int, link *goquery.Selection) {
		rawHref := strings.TrimSpace(link.AttrOr("href", ""))
		text := normalizeWikiText(link.Text())
		href := normalizeWikiHref(rawHref)
		if href == "" || text == "" || seen[href] {
			return
		}
		seen[href] = true
		links = append(links, RecentUpdateLink{Text: text, URL: href})
	})
	return links
}

func normalizeWikiHref(raw string) string {
	if raw == "" || strings.HasPrefix(raw, "#") {
		return ""
	}
	if strings.HasPrefix(raw, "/w/") {
		return "https://anothereden.wiki" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host != "anothereden.wiki" || !strings.HasPrefix(parsed.Path, "/w/") {
		return ""
	}
	return parsed.String()
}

func normalizeWikiText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
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

func getCharacterNameURL(doc *goquery.Document, wikiURL string) string {
	wikiTitle := strings.TrimPrefix(wikiURL, constants.AEWIKI_BASE_URL)
	wikiName := characterNameFromWikiTitle(wikiTitle)
	if logic.ResolveSpoilerName(wikiName) != wikiName {
		return wikiURL
	}
	aliasName := logic.ResolveForDB(wikiName)
	if aliasName != wikiName {
		return constants.AEWIKI_BASE_URL + strings.ReplaceAll(aliasName, " ", "_")
	}
	return getRedirectURL(doc, wikiURL)
}

func characterNameFromWikiTitle(title string) string {
	name := title
	if idx := strings.Index(name, "("); idx != -1 {
		name = name[:idx]
	}
	name = strings.TrimSpace(strings.ReplaceAll(name, "_", " "))
	if strings.Contains(title, constants.AEWIKI_ALTER_SUFFIX) && !strings.HasSuffix(name, " (Alter)") {
		name += " (Alter)"
	}
	return name
}

func ExtractCharacterInfoFromAEWikiDoc(doc *goquery.Document, wikiURL string) (*types.CharacterInfoFromAEWikiURL, error) {
	redirectURL := getCharacterNameURL(doc, wikiURL)

	title := strings.TrimPrefix(redirectURL, "https://anothereden.wiki/w/")
	info := &types.CharacterInfoFromAEWikiURL{}

	info.EnglishName = characterNameFromWikiTitle(title)

	spoilerName := info.EnglishName
	aliasName := logic.ResolveForDB(info.EnglishName)
	if aliasName != info.EnglishName {
		info.SpoilerEnglishName = info.EnglishName
		info.EnglishName = aliasName
	} else if resolvedName := logic.ResolveSpoilerName(info.EnglishName); resolvedName != info.EnglishName {
		info.SpoilerEnglishName = resolvedName
		spoilerName = resolvedName
	}

	// 이시층 여부 확인
	isAlter := strings.Contains(title, constants.AEWIKI_ALTER_SUFFIX) || strings.HasSuffix(spoilerName, " (Alter)")
	info.IsAlter = isAlter

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

func GetCharacterInfo(doc *goquery.Document, wikiURL string) (*types.CharacterInfoFromAEWiki, error) {
	baseInfo, err := ExtractCharacterInfoFromAEWikiDoc(doc, wikiURL)
	if err != nil {
		return nil, fmt.Errorf("GetCharacterInfo: %w", err)
	}

	info := &types.CharacterInfoFromAEWiki{
		CharacterInfoFromAEWikiURL: *baseInfo,
		WikiURL:                    getRedirectURL(doc, wikiURL),
	}

	generalData := doc.Find("article[title='General Data'] td")

	rarityText := generalData.Eq(0).Text()
	info.IsAwaken = logic.CheckAwaken(rarityText)
	info.HasFourStar = logic.HasFourStar(rarityText)
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

	otherDataArticle := doc.Find("article[title='Other Data']")
	otherData := otherDataArticle.Find("td")
	code := 0
	fmt.Sscanf(strings.TrimSpace(otherData.Eq(1).Text()), "%d", &code)
	info.GameID = code

	updateDateStr, err := extractUpdateDate(otherDataArticle)
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
			return nil, fmt.Errorf("no manifest weapon link found")
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

	bookEndpoint := strings.TrimSpace(characterClassTable.Eq(7).Find("a").Eq(0).AttrOr("href", ""))
	bookLink := ""
	if bookEndpoint != "" {
		bookLink = "https://anothereden.wiki" + bookEndpoint
	}

	className := strings.TrimSpace(characterClassTable.Eq(7).Text())
	// Remove newlines and control characters
	className = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, className)
	// Remove trailing " ...▽ ..." or " ... x1" suffixes
	if idx := strings.Index(className, " ..."); idx != -1 {
		className = strings.TrimSpace(className[:idx])
	}
	isFreeClass := strings.Contains(strings.ToLower(className), "(free)")
	// Remove (free), (paid) etc.
	if idx := strings.Index(className, "("); idx != -1 {
		className = strings.TrimSpace(className[:idx])
	}
	info.EnglishClassName = className
	info.Dungeons, err = getDungeonsFromAEWiki(info.Style, info.IsAlter, isFreeClass, bookLink)
	if err != nil {
		return nil, fmt.Errorf("get dungeons for %s (%s): %w", info.EnglishName, className, err)
	}

	return info, nil
}

func extractUpdateDate(otherDataArticle *goquery.Selection) (string, error) {
	for _, candidate := range updateDateCandidates(otherDataArticle) {
		updateDate, err := logic.ExtractDate(candidate)
		if err == nil {
			return updateDate, nil
		}
	}
	return "", fmt.Errorf("no update date found in Other Data")
}

func updateDateCandidates(otherDataArticle *goquery.Selection) []string {
	candidates := []string{}
	otherDataArticle.Find("tr").Each(func(i int, s *goquery.Selection) {
		text := normalizeWikiText(s.Text())
		if isUpdateDateRow(text) {
			candidates = append(candidates, text)
		}
	})
	otherDataArticle.Find("td").Each(func(i int, s *goquery.Selection) {
		if text := normalizeWikiText(s.Text()); text != "" {
			candidates = append(candidates, text)
		}
	})
	return candidates
}

func isUpdateDateRow(text string) bool {
	lowerText := strings.ToLower(text)
	return strings.Contains(lowerText, "update") || strings.Contains(lowerText, "release")
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

func getDungeonsFromAEWiki(style types.AEStyle, IsAlter bool, isFreeClass bool, bookLink string) ([]string, error) {
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

	if isFreeClass {
		return []string{"In-game"}, nil
	}

	if bookLink == "" {
		return nil, fmt.Errorf("missing class book link")
	}
	if !strings.HasPrefix(bookLink, constants.AEWIKI_BASE_URL) {
		return nil, fmt.Errorf("invalid class book link: %s", bookLink)
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

func GetBuddyInfoFromAEWiki(doc *goquery.Document, wikiURL string) (*types.BuddyInfoFromAEWiki, error) {
	var err error

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
			return nil, fmt.Errorf("GetBuddyInfoFromAEWiki: %w", err)
		}
		info.PartnerLink = getRedirectURL(partnerDoc, partnerURL)
	} else {
		info.PartnerLink = "None"
	}

	return info, nil
}
