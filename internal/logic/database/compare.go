package database

import (
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
)

// CompareDungeon compares dungeon data from wiki with database JSONB
func (s *Service) CompareDungeon(id int, info types.CharacterInfoFromAEWiki) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	common.Log.Info("Dungeon Mappings", logger.Field{Key: "characterID", Value: characterID})

	jsonbData, err := s.queries.GetCharacterJSONBData(ctx, characterID)
	if err != nil {
		common.Log.Warn("Character not found in database - New dungeons from Wiki", logger.Field{Key: "dungeons", Value: info.Dungeons})
		return
	}

	type DungeonLinks struct {
		AewikiURL string `json:"aewikiURL"`
		AltemaURL string `json:"altemaURL"`
	}

	type DungeonEntry struct {
		ID          string       `json:"id"`
		Links       DungeonLinks `json:"links"`
		Description *string      `json:"description"`
	}

	var dungeonEntries []DungeonEntry
	if len(jsonbData.DungeonsData) > 0 {
		if err := json.Unmarshal(jsonbData.DungeonsData, &dungeonEntries); err != nil {
			common.Log.Error("Error parsing dungeons_data", logger.Field{Key: "error", Value: err})
			return
		}
	}

	existingMappings := make(map[string]bool)
	dbDungeons := []string{}
	for _, d := range dungeonEntries {
		trans, err := s.queries.GetTranslation(ctx, d.ID)
		if err == nil {
			existingMappings[trans.En] = true
			dbDungeons = append(dbDungeons, trans.En)
		}
	}

	fmt.Printf("  DB Dungeons:   %v\n", dbDungeons)
	fmt.Printf("  Wiki Dungeons: %v\n", info.Dungeons)

	for _, dungeon := range info.Dungeons {
		if _, exists := existingMappings[dungeon]; exists {
			fmt.Printf("  %s✓%s %s (exists in DB)\n", common.ColorGreen, common.ColorReset, dungeon)
		} else {
			fmt.Printf("  %s+ %s (new in Wiki)%s\n", common.ColorRed, dungeon, common.ColorReset)
		}
	}

	for englishName := range existingMappings {
		found := false
		for _, dungeon := range info.Dungeons {
			if dungeon == englishName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  %s- %s (removed from Wiki)%s\n", common.ColorRed, englishName, common.ColorReset)
		}
	}
	fmt.Println()
}

// ComparePersonality compares personality data from wiki with database JSONB
func (s *Service) ComparePersonality(id int, info types.CharacterInfoFromAEWiki) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	common.Log.Info("Personality Mappings", logger.Field{Key: "characterID", Value: characterID})

	jsonbData, err := s.queries.GetCharacterJSONBData(ctx, characterID)
	if err != nil {
		common.Log.Warn("Character not found in database - New personalities from Wiki", logger.Field{Key: "personalities", Value: info.Personalities})
		return
	}

	type PersonalityEntry struct {
		ID          string  `json:"id"`
		Description *string `json:"description"`
	}

	var personalityEntries []PersonalityEntry
	if len(jsonbData.PersonalitiesData) > 0 {
		if err := json.Unmarshal(jsonbData.PersonalitiesData, &personalityEntries); err != nil {
			common.Log.Error("Error parsing personalities_data", logger.Field{Key: "error", Value: err})
			return
		}
	}

	existingMappings := make(map[string]bool)
	dbPersonalities := []string{}
	for _, p := range personalityEntries {
		trans, err := s.queries.GetTranslation(ctx, p.ID)
		if err == nil {
			existingMappings[trans.En] = true
			dbPersonalities = append(dbPersonalities, trans.En)
		}
	}

	fmt.Printf("  DB Personalities:   %v\n", dbPersonalities)
	fmt.Printf("  Wiki Personalities: %v\n", info.Personalities)

	for _, personality := range info.Personalities {
		if _, exists := existingMappings[personality]; exists {
			fmt.Printf("  %s✓%s %s (exists in DB)\n", common.ColorGreen, common.ColorReset, personality)
		} else {
			fmt.Printf("  %s+ %s (new in Wiki)%s\n", common.ColorRed, personality, common.ColorReset)
		}
	}

	for englishName := range existingMappings {
		found := false
		for _, p := range info.Personalities {
			if p == englishName {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("  %s- %s (removed from Wiki)%s\n", common.ColorRed, englishName, common.ColorReset)
		}
	}
	fmt.Println()
}
