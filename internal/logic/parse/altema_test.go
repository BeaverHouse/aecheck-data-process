package parse

import (
	"testing"
)

// TestGetAltemaLink tests the GetAltemaLink function with predefined character names and links
func TestGetAltemaLink(t *testing.T) {
	// Define test cases with character names and expected links
	testCases := []struct {
		name     string
		expected string
		isTales  bool
	}{
		{
			name:     "義侠の猟人 ラクレア",
			expected: "https://altema.jp/anaden/chara/1154",
			isTales:  false,
		},
		{
			name:     "アナベル(ES)",
			expected: "https://altema.jp/anaden/chara/1077",
			isTales:  false,
		},
		{
			name:     "シオン",
			expected: "https://altema.jp/anaden/chara/5",
			isTales:  false,
		},
		{
			name:     "シオン",
			expected: "https://altema.jp/anaden/chara/985",
			isTales:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			link, err := GetAltemaLink(tc.name, tc.isTales)
			if err != nil {
				t.Errorf("GetAltemaLink(%q) returned error: %v", tc.name, err)
			}
			if link != tc.expected {
				t.Errorf("GetAltemaLink(%q) = %q, want %q", tc.name, link, tc.expected)
			}
		})
	}
}
