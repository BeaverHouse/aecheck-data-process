package parse

import (
	"testing"
	"time"

	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"

	"github.com/stretchr/testify/require"
)

var (
	CHARACTER_LINKS = []string{
		"https://anothereden.wiki/w/Linaria",
		"https://anothereden.wiki/w/Velette_(Another_Style)",
		"https://anothereden.wiki/w/Red_Clad_Flam._(Another_Style)",
		"https://anothereden.wiki/w/Selfless_Seeker",
	}
	BUDDY_LINKS = []string{
		"https://anothereden.wiki/w/Gunce_(Another_Style)",
	}
	TEST_CHARACTER_LINKS = []string{
		"https://anothereden.wiki/w/Red_Clad_Flam._(Another_Style)",
		"https://anothereden.wiki/w/Thillelille",
	}
	TEST_BUDDY_LINKS = []string{
		"https://anothereden.wiki/w/Gunce_(Another_Style)",
		"https://anothereden.wiki/w/Thys%C3%ADa",
		"https://anothereden.wiki/w/Limil",
	}
	// Initialize with embedded type
	EXPECTED_CHARACTER_RESULT = []types.CharacterInfoFromAEWiki{
		{
			CharacterInfoFromAEWikiURL: types.CharacterInfoFromAEWikiURL{
				EnglishName: "Dewey (Alter)",
				Style:       types.StyleAS,
				IsAlter:     true,
			},
			GameID:           101060141,
			EnglishClassName: "Last Haven",
			IsAwaken:         true,
			LightShadow:      types.LSShadow,
			Category:         types.AECategoryEncounter,
			Personalities:    []string{"Fists", "Gun", "Lost Laboratory", "Fire", "Crystal", "IDA School", "Hood"},
			UpdateDate:       "2025-05-08",
			MaxManifest:      0,
			IsManifestCustom: false,
			Dungeons:         []string{"Treatise"},
			WikiURL:          "https://anothereden.wiki/w/Dewey_(Alter)_(Another_Style)",
		},
		{
			CharacterInfoFromAEWikiURL: types.CharacterInfoFromAEWikiURL{
				EnglishName: "Thillelille",
				Style:       types.StyleNS,
				IsAlter:     false,
			},
			GameID:           101010121,
			EnglishClassName: "Abyssal Devotee",
			IsAwaken:         false,
			LightShadow:      types.LSShadow,
			Category:         types.AECategoryEncounter,
			Personalities:    []string{"Sweet tooth", "Clergy", "West", "Sword", "Fire", "Shade"},
			UpdateDate:       "2020-04-11",
			MaxManifest:      0,
			IsManifestCustom: false,
			Dungeons: []string{
				"Antiquity Zerberiya Continent: Crystal",
				"Antiquity Zerberiya Continent: Shade",
				"Antiquity Zerberiya Continent: Thunder",
				"City of Lost Paradise",
				"Nadara Volcano",
			},
			WikiURL: "https://anothereden.wiki/w/Thillelille",
		},
	}
	EXPECTED_BUDDY_RESULT = []types.BuddyInfoFromAEWiki{
		{
			GameID:      2000000009,
			EnglishName: "Gunce",
			Style:       types.StyleAS,
			PartnerLink: "https://anothereden.wiki/w/Velette_(Another_Style)",
			WikiURL:     "https://anothereden.wiki/w/Gunce_(Another_Style)",
		},
		{
			GameID:      2000000014,
			EnglishName: "Thysía",
			Style:       types.StyleNS,
			PartnerLink: "None",
			WikiURL:     "https://anothereden.wiki/w/Thys%C3%ADa",
		},
		{
			GameID:      2000000008,
			EnglishName: "Limil",
			Style:       types.StyleNS,
			PartnerLink: "https://anothereden.wiki/w/Ilulu_(Alter)",
			WikiURL:     "https://anothereden.wiki/w/Limil",
		},
	}
)

func init() {
	common.InitLogger()
}

func TestGetRecentLinks(t *testing.T) {
	characterLinks, err := GetRecentLinks(time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), time.Now(), false)
	require.NoError(t, err)
	require.ElementsMatch(t, CHARACTER_LINKS, characterLinks)

	buddyLinks, err := GetRecentLinks(time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC), time.Now(), true)
	require.NoError(t, err)
	require.ElementsMatch(t, BUDDY_LINKS, buddyLinks)
}

func TestGetCharacterInfo(t *testing.T) {
	for i := range TEST_CHARACTER_LINKS {
		result, err := GetCharacterInfo(TEST_CHARACTER_LINKS[i])
		require.NoError(t, err)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].GameID, result.GameID)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].EnglishName, result.EnglishName)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].EnglishClassName, result.EnglishClassName)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].Style, result.Style)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].IsAlter, result.IsAlter)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].IsAwaken, result.IsAwaken)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].LightShadow, result.LightShadow)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].Category, result.Category)
		require.ElementsMatch(t, EXPECTED_CHARACTER_RESULT[i].Personalities, result.Personalities)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].UpdateDate, result.UpdateDate)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].MaxManifest, result.MaxManifest)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].IsManifestCustom, result.IsManifestCustom)
		require.ElementsMatch(t, EXPECTED_CHARACTER_RESULT[i].Dungeons, result.Dungeons)
		require.Equal(t, EXPECTED_CHARACTER_RESULT[i].WikiURL, result.WikiURL)
	}
}

func TestGetBuddyInfo(t *testing.T) {
	for i := range TEST_BUDDY_LINKS {
		result, err := GetBuddyInfoFromAEWiki(TEST_BUDDY_LINKS[i])
		require.NoError(t, err)
		require.Equal(t, EXPECTED_BUDDY_RESULT[i].GameID, result.GameID)
		require.Equal(t, EXPECTED_BUDDY_RESULT[i].EnglishName, result.EnglishName)
		require.Equal(t, EXPECTED_BUDDY_RESULT[i].Style, result.Style)
		require.Equal(t, EXPECTED_BUDDY_RESULT[i].PartnerLink, result.PartnerLink)
		require.Equal(t, EXPECTED_BUDDY_RESULT[i].WikiURL, result.WikiURL)
	}
}
