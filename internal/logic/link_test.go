package logic

import (
	"aecheck-data-process/internal/types"
	"testing"
)

func TestFindSeesaaLink(t *testing.T) {
	tests := []struct {
		info         types.CharacterInfoFromAEWiki
		japaneseName string
		expected     string
	}{
		{
			info: types.CharacterInfoFromAEWiki{
				CharacterInfoFromAEWikiURL: types.CharacterInfoFromAEWikiURL{
					Style: types.StyleAS,
				},
			},
			japaneseName: "赤套の炎使い",
			expected:     "https://anothereden.game-info.wiki/d/%C0%D6%C5%E5%A4%CE%B1%EA%BB%C8%A4%A4%28%A5%A2%A5%CA%A5%B6%A1%BC%A5%B9%A5%BF%A5%A4%A5%EB%29",
		},
		{
			info: types.CharacterInfoFromAEWiki{
				CharacterInfoFromAEWikiURL: types.CharacterInfoFromAEWikiURL{
					Style: types.StyleNS,
				},
			},
			japaneseName: "ニルヤ",
			expected:     "https://anothereden.game-info.wiki/d/%A5%CB%A5%EB%A5%E4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.japaneseName, func(t *testing.T) {
			got := FindSeesaaLink(tt.info, tt.japaneseName)
			if got != tt.expected {
				t.Errorf("FindSeesaaLink(%+v, %q) = %v, want %v", tt.info, tt.japaneseName, got, tt.expected)
			}
		})
	}
}
