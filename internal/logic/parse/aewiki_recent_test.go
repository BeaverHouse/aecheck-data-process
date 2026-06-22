package parse

import "testing"

func TestExtractRecentUpdateGroups(t *testing.T) {
	html := []byte(`
<div id="main-page-new-content">
  <div class="char-time-block">
    <div>Released: <b>Jun 18</b></div>
    <a href="/w/Xianhua_(Another_Style)">Xianhua AS</a>
    <a href="/w/Havoc">Havoc</a>
  </div>
  <div class="char-time-block">
    <div class="char-pics">
      <div>Sidekick: Jun 18</div>
      <div><a href="/w/Choco">Choco</a></div>
    </div>
  </div>
  <div class="char-time-block">
    <div class="char-pics">
      <div>New Stellar Awakening: Jun 18</div>
      <div><a href="/w/Amy">Amy</a><a href="/w/Amy">Amy</a></div>
      <div><a href="https://anothereden.wiki/w/Mistrare">Mistrare</a></div>
    </div>
  </div>
</div>`)

	groups, err := ExtractRecentUpdateGroups(html)
	if err != nil {
		t.Fatalf("ExtractRecentUpdateGroups() error = %v", err)
	}
	if len(groups) != 3 {
		t.Fatalf("len(groups) = %d, want 3", len(groups))
	}

	assertGroup(t, groups[0], RecentUpdateReleased, []string{
		"https://anothereden.wiki/w/Xianhua_(Another_Style)",
		"https://anothereden.wiki/w/Havoc",
	})
	assertGroup(t, groups[1], RecentUpdateSidekick, []string{
		"https://anothereden.wiki/w/Choco",
	})
	assertGroup(t, groups[2], RecentUpdateStellar, []string{
		"https://anothereden.wiki/w/Amy",
		"https://anothereden.wiki/w/Mistrare",
	})
}

func assertGroup(t *testing.T, group RecentUpdateGroup, wantKind RecentUpdateKind, wantURLs []string) {
	t.Helper()

	if group.Kind != wantKind {
		t.Fatalf("group.Kind = %q, want %q", group.Kind, wantKind)
	}
	if len(group.Links) != len(wantURLs) {
		t.Fatalf("len(group.Links) = %d, want %d", len(group.Links), len(wantURLs))
	}
	for i, wantURL := range wantURLs {
		if group.Links[i].URL != wantURL {
			t.Fatalf("group.Links[%d].URL = %q, want %q", i, group.Links[i].URL, wantURL)
		}
	}
}
