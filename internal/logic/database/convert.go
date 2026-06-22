package database

import (
	"aecheck-data-process/internal/logic/common"
	"context"
	"encoding/json"

	"github.com/BeaverHouse/go-common/logger"
)

func (s *Service) convertPersonalitiesToJSONB(ctx context.Context, personalities []string) ([]byte, error) {
	type PersonalityEntry struct {
		ID          string  `json:"id"`
		Description *string `json:"description"`
	}

	entries := []PersonalityEntry{}
	for _, name := range personalities {
		personalityID, err := s.queries.GetKeyByEnglishName(ctx, name)
		if err != nil {
			common.Log.Warn("Personality not found in translations", logger.Field{Key: "name", Value: name})
			continue
		}
		entries = append(entries, PersonalityEntry{
			ID:          personalityID,
			Description: nil,
		})
	}

	if len(entries) == 0 {
		return json.Marshal([]PersonalityEntry{})
	}
	return json.Marshal(entries)
}

func (s *Service) convertDungeonsToJSONB(ctx context.Context, dungeons []string) ([]byte, error) {
	type DungeonLinks struct {
		AewikiURL string `json:"aewikiURL"`
		AltemaURL string `json:"altemaURL"`
	}

	type DungeonEntry struct {
		ID          string       `json:"id"`
		Links       DungeonLinks `json:"links"`
		Description *string      `json:"description"`
	}

	entries := []DungeonEntry{}
	for _, name := range dungeons {
		dungeonID, err := s.queries.GetKeyByEnglishName(ctx, name)
		if err != nil {
			common.Log.Warn("Dungeon not found in translations", logger.Field{Key: "name", Value: name})
			continue
		}

		dungeon, err := s.queries.GetDungeonByID(ctx, dungeonID)
		if err != nil {
			common.Log.Warn("Dungeon details not found", logger.Field{Key: "dungeonID", Value: dungeonID})
			continue
		}

		aewikiURL := ""
		if dungeon.AewikiUrl.Valid {
			aewikiURL = dungeon.AewikiUrl.String
		}

		altemaURL := ""
		if dungeon.AltemaUrl.Valid {
			altemaURL = dungeon.AltemaUrl.String
		}

		entries = append(entries, DungeonEntry{
			ID: dungeonID,
			Links: DungeonLinks{
				AewikiURL: aewikiURL,
				AltemaURL: altemaURL,
			},
			Description: nil,
		})
	}

	if len(entries) == 0 {
		return json.Marshal([]DungeonEntry{})
	}
	return json.Marshal(entries)
}
