package parse

import (
	"aecheck-data-process/internal/types"
	"testing"
)

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
