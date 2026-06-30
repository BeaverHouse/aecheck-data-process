package parse

import (
	"aecheck-data-process/internal/types"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestExtractCharacterInfoKeepsSpoilerAliasAsEnglishName(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<html>
  <head><link rel="canonical" href="https://anothereden.wiki/w/Spirika"></head>
</html>`))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	info, err := ExtractCharacterInfoFromAEWikiDoc(doc, "https://anothereden.wiki/w/Chelika")
	if err != nil {
		t.Fatalf("ExtractCharacterInfoFromAEWikiDoc() error = %v", err)
	}
	if info.EnglishName != "Chelika" {
		t.Fatalf("EnglishName = %q, want Chelika", info.EnglishName)
	}
	if info.SpoilerEnglishName != "Spirika" {
		t.Fatalf("SpoilerEnglishName = %q, want Spirika", info.SpoilerEnglishName)
	}
	if got := getCharacterNameURL(doc, "https://anothereden.wiki/w/Chelika"); got != "https://anothereden.wiki/w/Chelika" {
		t.Fatalf("getCharacterNameURL() = %q, want alias URL", got)
	}
}

func TestExtractCharacterInfoMapsSpoilerTrueNameURLToAlias(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<html>
  <head><link rel="canonical" href="https://anothereden.wiki/w/Spirika"></head>
</html>`))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	info, err := ExtractCharacterInfoFromAEWikiDoc(doc, "https://anothereden.wiki/w/Spirika")
	if err != nil {
		t.Fatalf("ExtractCharacterInfoFromAEWikiDoc() error = %v", err)
	}
	if info.EnglishName != "Chelika" {
		t.Fatalf("EnglishName = %q, want Chelika", info.EnglishName)
	}
	if info.SpoilerEnglishName != "Spirika" {
		t.Fatalf("SpoilerEnglishName = %q, want Spirika", info.SpoilerEnglishName)
	}
	if got := getCharacterNameURL(doc, "https://anothereden.wiki/w/Spirika"); got != "https://anothereden.wiki/w/Chelika" {
		t.Fatalf("getCharacterNameURL() = %q, want alias URL", got)
	}
}

func TestExtractCharacterInfoMapsAlterTrueNameURLToAlias(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<html>
  <head><link rel="canonical" href="https://anothereden.wiki/w/Dewey_(Alter)"></head>
</html>`))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	info, err := ExtractCharacterInfoFromAEWikiDoc(doc, "https://anothereden.wiki/w/Dewey_(Alter)")
	if err != nil {
		t.Fatalf("ExtractCharacterInfoFromAEWikiDoc() error = %v", err)
	}
	if info.EnglishName != "Red Clad Flam." {
		t.Fatalf("EnglishName = %q, want Red Clad Flam.", info.EnglishName)
	}
	if info.SpoilerEnglishName != "Dewey (Alter)" {
		t.Fatalf("SpoilerEnglishName = %q, want Dewey (Alter)", info.SpoilerEnglishName)
	}
	if !info.IsAlter {
		t.Fatal("IsAlter = false, want true")
	}
	if got := getCharacterNameURL(doc, "https://anothereden.wiki/w/Dewey_(Alter)"); got != "https://anothereden.wiki/w/Red_Clad_Flam." {
		t.Fatalf("getCharacterNameURL() = %q, want alias URL", got)
	}
}

func TestGetCharacterInfoKeepsRedirectWikiURL(t *testing.T) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(`
<html>
  <head><link rel="canonical" href="https://anothereden.wiki/w/Dewey_(Alter)"></head>
  <article title="General Data"><table><tr>
    <td>5★</td><td>Fire</td><td></td><td></td><td></td><td>light</td><td>Encounter</td><td></td>
  </tr></table></article>
  <article title="Other Data"><table><tr><td>Character ID</td><td>1</td></tr><tr><td>Release Date</td><td>January 1, 2026</td></tr></table></article>
  <div class="character-weapon"><table><tr></tr><tr><th colspan="0"></th></tr></table></div>
  <div class="character-class"><table><tr>
    <td></td><td></td><td></td><td></td><td></td><td></td><td></td><td>Red Clad Flam. (free)</td>
  </tr></table></div>
</html>`))
	if err != nil {
		t.Fatalf("NewDocumentFromReader() error = %v", err)
	}

	info, err := GetCharacterInfo(doc, "https://anothereden.wiki/w/Dewey_(Alter)")
	if err != nil {
		t.Fatalf("GetCharacterInfo() error = %v", err)
	}
	if info.WikiURL != "https://anothereden.wiki/w/Dewey_(Alter)" {
		t.Fatalf("WikiURL = %q, want redirect URL", info.WikiURL)
	}
}

func TestGetDungeonsFromAEWikiMissingBookLink(t *testing.T) {
	_, err := getDungeonsFromAEWiki(types.StyleNS, false, false, "")
	if err == nil {
		t.Fatal("getDungeonsFromAEWiki() error = nil, want missing class book link error")
	}
}

func TestGetDungeonsFromAEWikiSkipsBookLinkForFreeClass(t *testing.T) {
	got, err := getDungeonsFromAEWiki(types.StyleNS, false, true, "")
	if err != nil {
		t.Fatalf("getDungeonsFromAEWiki() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0] != "In-game" {
		t.Fatalf("getDungeonsFromAEWiki() = %v, want [In-game]", got)
	}
}

func TestGetDungeonsFromAEWikiSkipsBookLinkForAlter(t *testing.T) {
	got, err := getDungeonsFromAEWiki(types.StyleNS, true, false, "")
	if err != nil {
		t.Fatalf("getDungeonsFromAEWiki() error = %v, want nil", err)
	}
	if len(got) != 1 || got[0] != "Opus" {
		t.Fatalf("getDungeonsFromAEWiki() = %v, want [Opus]", got)
	}
}
