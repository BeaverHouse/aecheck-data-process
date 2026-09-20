package parse

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestBuddyGameIDReadsIconAltAndFileName(t *testing.T) {
	tests := map[string]struct {
		html string
		want int
	}{
		"legacy alt text": {
			html: `<div class="sidekick-icon"><img alt="2000000013 command" src="/thumb/80px-2000000013_s2_command.png"></div>`,
			want: 2000000013,
		},
		"current file name": {
			html: `<div class="sidekick-icon"><img src="https://cdn.example/images/thumb/b/b0/2000000013_s2_command.png/80px-2000000013_s2_command.png"></div>`,
			want: 2000000013,
		},
		"no icon": {
			html: `<div class="sidekick-icon"></div>`,
			want: 0,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(test.html))
			if err != nil {
				t.Fatalf("NewDocumentFromReader() error = %v", err)
			}

			if got := buddyGameID(doc.Find("div.sidekick-icon img")); got != test.want {
				t.Fatalf("buddyGameID() = %d, want %d", got, test.want)
			}
		})
	}
}

const currentTabberOtherData = `
<html><body>
  <div class="tabber">
    <div class="tabber__panel" id="tabber-Other_Data">
      <table><tbody>
        <tr><th>Id</th><td>101000451</td></tr>
        <tr><th>Release GL Date / JP Date</th><td>Sep 17, 2026 / Sep 17, 2026</td></tr>
      </tbody></table>
    </div>
  </div>
</body></html>`

const legacyTabberOtherData = `
<html><body>
  <article class="tabber__panel" title="Other Data">
    <table><tbody>
      <tr><th>Id</th><td>101000451</td></tr>
      <tr><th>Release GL Date / JP Date</th><td>Sep 17, 2026 / Sep 17, 2026</td></tr>
    </tbody></table>
  </article>
</body></html>`

func TestExtractUpdateDateReadsBothTabberMarkups(t *testing.T) {
	tests := map[string]string{
		"current": currentTabberOtherData,
		"legacy":  legacyTabberOtherData,
	}

	for name, html := range tests {
		t.Run(name, func(t *testing.T) {
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("NewDocumentFromReader() error = %v", err)
			}

			panel := findInfoPanel(doc, "Other Data")
			got, err := extractUpdateDate(panel)
			if err != nil {
				t.Fatalf("extractUpdateDate() error = %v", err)
			}
			if got != "2026-09-17" {
				t.Fatalf("extractUpdateDate() = %q, want 2026-09-17", got)
			}
			if id := strings.TrimSpace(panel.Find("td").Eq(0).Text()); id != "101000451" {
				t.Fatalf("game id cell = %q, want 101000451", id)
			}
		})
	}
}
