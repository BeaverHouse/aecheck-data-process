package database

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service wraps sqlc queries to provide the same interface as app/database
type Service struct {
	pool    *pgxpool.Pool
	queries *postgres.Queries
}

// NewService creates a new database service
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		queries: postgres.New(pool),
	}
}

// CompareCharacter compares character info from wiki with database
func (s *Service) CompareCharacter(info types.CharacterInfoFromAEWiki, scrapedSeesaaURL string, id int) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	fmt.Println()
	fmt.Printf("=== Comparing Character: %s ===\n", characterID)
	fmt.Println()

	// Get character from database
	dbChar, err := s.queries.GetCharacterWithTranslation(ctx, characterID)
	if err != nil {
		fmt.Printf("%sCharacter %s not found in database - New character from Wiki:%s\n", common.ColorRed, characterID, common.ColorReset)

		// Display Wiki data for new character
		calculatedAlterCharacterID := pgtype.Text{Valid: false}
		calculatedAlterLink := logic.FindAlterLink(info)
		if calculatedAlterLink != "" {
			alterCode, err := s.queries.GetCharacterCodeByWikiURL(ctx, pgtype.Text{String: calculatedAlterLink, Valid: true})
			if err == nil {
				calculatedAlterCharacterID = pgtype.Text{String: alterCode, Valid: true}
			}
		}

		wikiAlterChar := "NULL"
		if calculatedAlterCharacterID.Valid {
			wikiAlterChar = calculatedAlterCharacterID.String
		}

		fmt.Printf("  %-20s: %s%s%s\n", "EnglishName", common.ColorGreen, info.EnglishName, common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "Style", common.ColorGreen, string(info.Style), common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "LightShadow", common.ColorGreen, string(info.LightShadow), common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "Category", common.ColorGreen, string(info.Category), common.ColorReset)
		fmt.Printf("  %-20s: %s%d%s\n", "MaxManifest", common.ColorGreen, info.MaxManifest, common.ColorReset)
		fmt.Printf("  %-20s: %s%v%s\n", "IsAwaken", common.ColorGreen, info.IsAwaken, common.ColorReset)
		fmt.Printf("  %-20s: %s%v%s\n", "IsAlter", common.ColorGreen, info.IsAlter, common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "AlterCharacter", common.ColorGreen, wikiAlterChar, common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "WikiURL", common.ColorGreen, info.WikiURL, common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "SeesaaURL", common.ColorGreen, scrapedSeesaaURL, common.ColorReset)
		fmt.Printf("  %-20s: %s%s%s\n", "UpdateDate", common.ColorGreen, info.UpdateDate, common.ColorReset)
		fmt.Printf("  %-20s: %s%v%s\n", "CustomManifest", common.ColorGreen, info.IsManifestCustom, common.ColorReset)
		fmt.Printf("  %-20s: %s%v%s\n", "Personalities", common.ColorGreen, info.Personalities, common.ColorReset)
		fmt.Printf("  %-20s: %s%v%s\n", "Dungeons", common.ColorGreen, info.Dungeons, common.ColorReset)
		fmt.Println()
		return
	}

	// Get alter character code if exists
	calculatedAlterCharacterID := pgtype.Text{Valid: false}
	calculatedAlterLink := logic.FindAlterLink(info)
	if calculatedAlterLink != "" {
		alterCode, err := s.queries.GetCharacterCodeByWikiURL(ctx, pgtype.Text{String: calculatedAlterLink, Valid: true})
		if err == nil {
			calculatedAlterCharacterID = pgtype.Text{String: alterCode, Valid: true}
		}
	}

	// Compare all fields and display in color
	compareField := func(name string, dbVal, wikiVal interface{}) {
		dbStr := fmt.Sprintf("%v", dbVal)
		wikiStr := fmt.Sprintf("%v", wikiVal)

		if dbStr == wikiStr {
			// Same - Green
			fmt.Printf("  %-20s: %s%-30s%s (DB) = %s%-30s%s (Wiki)\n",
				name, common.ColorGreen, dbStr, common.ColorReset, common.ColorGreen, wikiStr, common.ColorReset)
		} else {
			// Different - Red
			fmt.Printf("  %-20s: %s%-30s%s (DB) ≠ %s%-30s%s (Wiki)\n",
				name, common.ColorRed, dbStr, common.ColorReset, common.ColorRed, wikiStr, common.ColorReset)
		}
	}

	// Handle EnglishName with INITIAL_AC_NAMES check
	dbEnglishName := dbChar.EnglishName.String
	wikiEnglishName := info.EnglishName
	if constants.INITIAL_AC_NAMES[dbEnglishName] != "" {
		dbEnglishName = fmt.Sprintf("%s (aka %s)", dbEnglishName, constants.INITIAL_AC_NAMES[dbEnglishName])
	}
	compareField("EnglishName", dbEnglishName, wikiEnglishName)

	compareField("Style", dbChar.Style, string(info.Style))
	compareField("LightShadow", dbChar.LightShadow, string(info.LightShadow))
	compareField("Category", dbChar.Category, string(info.Category))
	compareField("MaxManifest", dbChar.MaxManifest, info.MaxManifest)
	compareField("IsAwaken", dbChar.IsAwaken, info.IsAwaken)
	compareField("IsAlter", dbChar.IsAlter, info.IsAlter)

	dbAlterChar := "NULL"
	if dbChar.AlterCharacter.Valid {
		dbAlterChar = dbChar.AlterCharacter.String
	}
	wikiAlterChar := "NULL"
	if calculatedAlterCharacterID.Valid {
		wikiAlterChar = calculatedAlterCharacterID.String
	}
	compareField("AlterCharacter", dbAlterChar, wikiAlterChar)

	compareField("WikiURL", dbChar.AewikiUrl.String, info.WikiURL)
	compareField("SeesaaURL", dbChar.SeesaaUrl.String, scrapedSeesaaURL)
	compareField("UpdateDate", dbChar.UpdateDate, info.UpdateDate)

	dbCustomManifest := false
	if dbChar.CustomManifest.Valid {
		dbCustomManifest = dbChar.CustomManifest.Bool
	}
	compareField("CustomManifest", dbCustomManifest, info.IsManifestCustom)

	fmt.Println()
}

// UpsertCharacter inserts or updates a character
func (s *Service) UpsertCharacter(info types.CharacterInfoFromAEWiki, seesaaURL string, id int, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] UpsertCharacter: %s (id=%d)\n", info.EnglishName, id)
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	// Get alter character code if exists
	calculatedAlterCharacterID := pgtype.Text{Valid: false}
	calculatedAlterLink := logic.FindAlterLink(info)
	if calculatedAlterLink != "" {
		alterCode, err := s.queries.GetCharacterCodeByWikiURL(ctx, pgtype.Text{String: calculatedAlterLink, Valid: true})
		if err == nil {
			calculatedAlterCharacterID = pgtype.Text{String: alterCode, Valid: true}
		}
	}

	// Parse update date
	updateDate := pgtype.Date{Valid: false}
	if info.UpdateDate != "" {
		t, err := time.Parse("2006-01-02", info.UpdateDate)
		if err == nil {
			updateDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	// Convert personalities to JSONB
	personalitiesData, err := s.convertPersonalitiesToJSONB(ctx, info.Personalities)
	if err != nil {
		panic(err)
	}

	// Convert dungeons to JSONB
	dungeonsData, err := s.convertDungeonsToJSONB(ctx, info.Dungeons)
	if err != nil {
		panic(err)
	}

	// Buddy data is nil for now (will be set separately if needed)
	var buddyData []byte = nil

	// Check if character exists
	exists, err := s.queries.CheckCharacterExists(ctx, characterID)
	if err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Character already exists: ", id)
		err = s.queries.UpdateCharacter(ctx, postgres.UpdateCharacterParams{
			CharacterCode:     fmt.Sprintf("c%d", info.GameID),
			Category:          string(info.Category),
			Style:             string(info.Style),
			LightShadow:       string(info.LightShadow),
			MaxManifest:       int32(info.MaxManifest),
			IsAwaken:          info.IsAwaken,
			IsAlter:           info.IsAlter,
			AlterCharacter:    calculatedAlterCharacterID,
			AewikiUrl:         pgtype.Text{String: info.WikiURL, Valid: true},
			SeesaaUrl:         pgtype.Text{String: seesaaURL, Valid: true},
			UpdateDate:        updateDate,
			CustomManifest:    pgtype.Bool{Bool: info.IsManifestCustom, Valid: true},
			PersonalitiesData: personalitiesData,
			DungeonsData:      dungeonsData,
			BuddyData:         buddyData,
			CharacterID:       characterID,
		})
		if err != nil {
			panic(err)
		}
	} else {
		err = s.queries.InsertCharacter(ctx, postgres.InsertCharacterParams{
			CharacterID:       characterID,
			CharacterCode:     fmt.Sprintf("c%d", info.GameID),
			Category:          string(info.Category),
			Style:             string(info.Style),
			LightShadow:       string(info.LightShadow),
			MaxManifest:       int32(info.MaxManifest),
			IsAwaken:          info.IsAwaken,
			IsAlter:           info.IsAlter,
			AlterCharacter:    calculatedAlterCharacterID,
			AewikiUrl:         pgtype.Text{String: info.WikiURL, Valid: true},
			SeesaaUrl:         pgtype.Text{String: seesaaURL, Valid: true},
			UpdateDate:        updateDate,
			CustomManifest:    pgtype.Bool{Bool: info.IsManifestCustom, Valid: true},
			PersonalitiesData: personalitiesData,
			DungeonsData:      dungeonsData,
			BuddyData:         buddyData,
		})
		if err != nil {
			panic(err)
		}
	}
}

// convertPersonalitiesToJSONB converts personality names to JSONB with IDs
func (s *Service) convertPersonalitiesToJSONB(ctx context.Context, personalities []string) ([]byte, error) {
	type PersonalityEntry struct {
		ID          string  `json:"id"`
		Description *string `json:"description"`
	}

	entries := []PersonalityEntry{}
	for _, name := range personalities {
		personalityID, err := s.queries.GetKeyByEnglishName(ctx, name)
		if err != nil {
			// If not found, skip or handle error
			fmt.Printf("Warning: Personality '%s' not found in translations\n", name)
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

// convertDungeonsToJSONB converts dungeon names to JSONB with full dungeon info
func (s *Service) convertDungeonsToJSONB(ctx context.Context, dungeons []string) ([]byte, error) {
	type DungeonLinks struct {
		AewikiURL string `json:"aewikiURL"`
		AltemaURL string `json:"altemaURL"`
	}

	type DungeonEntry struct {
		ID          string        `json:"id"`
		Links       DungeonLinks  `json:"links"`
		Description *string       `json:"description"`
	}

	entries := []DungeonEntry{}
	for _, name := range dungeons {
		dungeonID, err := s.queries.GetKeyByEnglishName(ctx, name)
		if err != nil {
			// If not found, skip or handle error
			fmt.Printf("Warning: Dungeon '%s' not found in translations\n", name)
			continue
		}

		// Get dungeon details from dungeons table
		dungeon, err := s.queries.GetDungeonByID(ctx, dungeonID)
		if err != nil {
			fmt.Printf("Warning: Dungeon details for '%s' not found\n", dungeonID)
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

// CheckFourStarUpdate checks if a four-star character needs updating
func (s *Service) CheckFourStarUpdate(info types.CharacterInfoFromAEWiki, excluded bool) (int, types.UpdateStatus) {
	if info.IsAlter {
		fmt.Println("Alter character starts with 5-star in default: ", info.EnglishName)
		return -1, types.NotNeeded
	} else if info.Style == types.StyleES {
		fmt.Println("ES character differs weapon type from default: ", info.EnglishName)
		return -1, types.NotNeeded
	} else if excluded {
		fmt.Println("Excluded character: ", info.EnglishName)
		return -1, types.NotNeeded
	}

	ctx := context.Background()
	characterCode := fmt.Sprintf("c%d", info.GameID)

	char, err := s.queries.GetFirstCharacterByCode(ctx, characterCode)
	if err != nil {
		// Character doesn't exist
		return -1, types.NotExists
	}

	id, err := strconv.Atoi(strings.Replace(char.CharacterID, "char", "", 1))
	if err != nil {
		panic(err)
	}

	if char.Style != "☆4" {
		return -1, types.NotNeeded
	} else if char.UpdatedAt.Valid && char.UpdatedAt.Time.After(time.Now().AddDate(0, -6, 0)) {
		return id, types.Updated
	} else {
		return id, types.NotUpdated
	}
}

// InsertFourStarCharacter inserts a four-star character by copying from existing ID
func (s *Service) InsertFourStarCharacter(id int, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] InsertFourStarCharacter: id=%d\n", id)
		return
	}

	ctx := context.Background()

	// Copy character to new ID
	err := s.queries.CopyCharacterToNewID(ctx, postgres.CopyCharacterToNewIDParams{
		CharacterID:   fmt.Sprintf("char%04d", id+1),
		CharacterID_2: fmt.Sprintf("char%04d", id),
	})
	if err != nil {
		panic(err)
	}

	// Update original to four-star
	err = s.queries.UpdateCharacterToFourStar(ctx, fmt.Sprintf("char%04d", id))
	if err != nil {
		panic(err)
	}
}

// UpdateFourStarCharacter updates a character to four-star
func (s *Service) UpdateFourStarCharacter(id int, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] UpdateFourStarCharacter: id=%d\n", id)
		return
	}

	ctx := context.Background()
	err := s.queries.UpdateCharacterToFourStar(ctx, fmt.Sprintf("char%04d", id))
	if err != nil {
		panic(err)
	}
}

// CompareDungeon compares dungeon data from wiki with database JSONB
func (s *Service) CompareDungeon(id int, info types.CharacterInfoFromAEWiki) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	fmt.Printf("=== Dungeon Mappings for %s ===\n", characterID)

	// Get JSONB data from character table
	jsonbData, err := s.queries.GetCharacterJSONBData(ctx, characterID)
	if err != nil {
		fmt.Printf("  %sCharacter not found in database - New dungeons from Wiki:%s\n", common.ColorRed, common.ColorReset)
		fmt.Printf("  %sWiki Dungeons: %v%s\n", common.ColorGreen, info.Dungeons, common.ColorReset)
		fmt.Println()
		return
	}

	type DungeonLinks struct {
		AewikiURL string `json:"aewikiURL"`
		AltemaURL string `json:"altemaURL"`
	}

	type DungeonEntry struct {
		ID          string        `json:"id"`
		Links       DungeonLinks  `json:"links"`
		Description *string       `json:"description"`
	}

	var dungeonEntries []DungeonEntry
	if len(jsonbData.DungeonsData) > 0 {
		if err := json.Unmarshal(jsonbData.DungeonsData, &dungeonEntries); err != nil {
			fmt.Printf("  Error parsing dungeons_data: %v\n", err)
			return
		}
	}

	// Build map of dungeon ID to english name from translations
	existingMappings := make(map[string]bool)
	dbDungeons := []string{}
	for _, d := range dungeonEntries {
		// Get English name from translation
		trans, err := s.queries.GetTranslation(ctx, d.ID)
		if err == nil {
			existingMappings[trans.En] = true
			dbDungeons = append(dbDungeons, trans.En)
		}
	}

	// Display comparison
	fmt.Printf("  DB Dungeons:   %v\n", dbDungeons)
	fmt.Printf("  Wiki Dungeons: %v\n", info.Dungeons)

	// Compare with new dungeon mappings
	for _, dungeon := range info.Dungeons {
		if _, exists := existingMappings[dungeon]; exists {
			fmt.Printf("  %s✓%s %s (exists in DB)\n", common.ColorGreen, common.ColorReset, dungeon)
		} else {
			fmt.Printf("  %s+ %s (new in Wiki)%s\n", common.ColorRed, dungeon, common.ColorReset)
		}
	}

	// Check for removed dungeons
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

// UpsertDungeon inserts or updates dungeon mappings
func (s *Service) UpsertDungeon(id int, info types.CharacterInfoFromAEWiki, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] UpsertDungeon: id=%d, dungeons=%v\n", id, info.Dungeons)
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	// First, soft delete all existing mappings
	err := s.queries.SoftDeleteDungeonMappings(ctx, characterID)
	if err != nil {
		panic(err)
	}

	// Insert new mappings
	fmt.Println("Dungeons: ", info.Dungeons)
	for _, dungeon := range info.Dungeons {
		dungeonID, err := s.queries.GetKeyByEnglishName(ctx, dungeon)
		if err != nil {
			panic(err)
		}

		err = s.queries.InsertDungeonMapping(ctx, postgres.InsertDungeonMappingParams{
			CharacterID: characterID,
			DungeonID:   dungeonID,
		})
		if err != nil {
			panic(err)
		}
	}
}

// PurgeDeletedDungeon removes soft-deleted dungeon mappings
func (s *Service) PurgeDeletedDungeon(id int, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] PurgeDeletedDungeon: id=%d\n", id)
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)
	err := s.queries.PurgeDeletedDungeonMappings(ctx, characterID)
	if err != nil {
		panic(err)
	}
}

// ComparePersonality compares personality data from wiki with database JSONB
func (s *Service) ComparePersonality(id int, info types.CharacterInfoFromAEWiki) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	fmt.Printf("=== Personality Mappings for %s ===\n", characterID)

	// Get JSONB data from character table
	jsonbData, err := s.queries.GetCharacterJSONBData(ctx, characterID)
	if err != nil {
		fmt.Printf("  %sCharacter not found in database - New personalities from Wiki:%s\n", common.ColorRed, common.ColorReset)
		fmt.Printf("  %sWiki Personalities: %v%s\n", common.ColorGreen, info.Personalities, common.ColorReset)
		fmt.Println()
		return
	}

	type PersonalityEntry struct {
		ID          string  `json:"id"`
		Description *string `json:"description"`
	}

	var personalityEntries []PersonalityEntry
	if len(jsonbData.PersonalitiesData) > 0 {
		if err := json.Unmarshal(jsonbData.PersonalitiesData, &personalityEntries); err != nil {
			fmt.Printf("  Error parsing personalities_data: %v\n", err)
			return
		}
	}

	// Build map of personality ID to english name from translations
	existingMappings := make(map[string]bool)
	dbPersonalities := []string{}
	for _, p := range personalityEntries {
		// Get English name from translation
		trans, err := s.queries.GetTranslation(ctx, p.ID)
		if err == nil {
			existingMappings[trans.En] = true
			dbPersonalities = append(dbPersonalities, trans.En)
		}
	}

	// Display comparison
	fmt.Printf("  DB Personalities:   %v\n", dbPersonalities)
	fmt.Printf("  Wiki Personalities: %v\n", info.Personalities)

	// Compare with new personality mappings
	for _, personality := range info.Personalities {
		if _, exists := existingMappings[personality]; exists {
			fmt.Printf("  %s✓%s %s (exists in DB)\n", common.ColorGreen, common.ColorReset, personality)
		} else {
			fmt.Printf("  %s+ %s (new in Wiki)%s\n", common.ColorRed, personality, common.ColorReset)
		}
	}

	// Check for removed personalities
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

// UpsertPersonality inserts or updates personality mappings
func (s *Service) UpsertPersonality(id int, info types.CharacterInfoFromAEWiki, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] UpsertPersonality: id=%d, personalities=%v\n", id, info.Personalities)
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	// First, soft delete all existing mappings
	err := s.queries.SoftDeletePersonalityMappings(ctx, characterID)
	if err != nil {
		panic(err)
	}

	// Insert new mappings
	fmt.Println("Personalities: ", info.Personalities)
	for _, personality := range info.Personalities {
		personalityID, err := s.queries.GetKeyByEnglishName(ctx, personality)
		if err != nil {
			panic(err)
		}

		err = s.queries.InsertPersonalityMapping(ctx, postgres.InsertPersonalityMappingParams{
			CharacterID:   characterID,
			PersonalityID: personalityID,
		})
		if err != nil {
			panic(err)
		}
	}
}

// PurgeDeletedPersonality removes soft-deleted personality mappings
func (s *Service) PurgeDeletedPersonality(id int, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] PurgeDeletedPersonality: id=%d\n", id)
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)
	err := s.queries.PurgeDeletedPersonalityMappings(ctx, characterID)
	if err != nil {
		panic(err)
	}
}

// CompareTranslations compares translation info with database
func (s *Service) CompareTranslations(info types.TranslationInfo, code string) {
	ctx := context.Background()

	fmt.Printf("=== Translation for %s ===\n", code)

	trans, err := s.queries.GetTranslation(ctx, code)
	if err != nil {
		fmt.Println("  Translation not found in database")
		fmt.Printf("  Wiki: KO=%s, EN=%s, JA=%s\n", info.KoreanName, info.EnglishName, info.JapaneseName)
		fmt.Println()
		return
	}

	compareField := func(name string, dbVal, wikiVal string) {
		if dbVal == wikiVal {
			fmt.Printf("  %-12s: %s%-20s%s (DB) = %s%-20s%s (Wiki)\n",
				name, common.ColorGreen, dbVal, common.ColorReset, common.ColorGreen, wikiVal, common.ColorReset)
		} else {
			fmt.Printf("  %-12s: %s%-20s%s (DB) ≠ %s%-20s%s (Wiki)\n",
				name, common.ColorRed, dbVal, common.ColorReset, common.ColorRed, wikiVal, common.ColorReset)
		}
	}

	compareField("Korean", trans.Ko, info.KoreanName)

	// Handle EnglishName with INITIAL_AC_NAMES check
	// Check if DB name matches Wiki name OR if the alter name matches Wiki name
	if trans.En == info.EnglishName || (constants.INITIAL_AC_NAMES[trans.En] != "" && constants.INITIAL_AC_NAMES[trans.En] == info.EnglishName) {
		// Match found
		if constants.INITIAL_AC_NAMES[trans.En] != "" {
			fmt.Printf("  %-12s: %s%-20s%s (DB) = %s%-20s%s (Wiki)\n",
				"English", common.ColorGreen, fmt.Sprintf("%s (aka %s)", trans.En, constants.INITIAL_AC_NAMES[trans.En]), common.ColorReset, common.ColorGreen, info.EnglishName, common.ColorReset)
		} else {
			compareField("English", trans.En, info.EnglishName)
		}
	} else {
		// No match
		if constants.INITIAL_AC_NAMES[trans.En] != "" {
			fmt.Printf("  %-12s: %s%-20s%s (DB) ≠ %s%-20s%s (Wiki)\n",
				"English", common.ColorRed, fmt.Sprintf("%s (aka %s)", trans.En, constants.INITIAL_AC_NAMES[trans.En]), common.ColorReset, common.ColorRed, info.EnglishName, common.ColorReset)
		} else {
			compareField("English", trans.En, info.EnglishName)
		}
	}

	compareField("Japanese", trans.Ja, info.JapaneseName)
	fmt.Println()
}

// UpsertTranslation inserts or updates translation
func (s *Service) UpsertTranslation(info types.TranslationInfo, code string, dryrun bool) {
	// Skip if this is an alter name
	for _, alterName := range constants.INITIAL_AC_NAMES {
		if alterName == info.EnglishName {
			return
		}
	}

	if dryrun {
		fmt.Printf("[DRYRUN] UpsertTranslation: code=%s, en=%s\n", code, info.EnglishName)
		return
	}

	ctx := context.Background()

	// Check if translation exists
	exists, err := s.queries.CheckTranslationExists(ctx, code)
	if err != nil {
		panic(err)
	}

	if exists {
		fmt.Println("Translation already exists: ", code)
		err = s.queries.UpdateTranslation(ctx, postgres.UpdateTranslationParams{
			Ko:  info.KoreanName,
			En:  info.EnglishName,
			Ja:  info.JapaneseName,
			Key: code,
		})
		if err != nil {
			panic(err)
		}
	} else {
		err = s.queries.InsertTranslation(ctx, postgres.InsertTranslationParams{
			Key: code,
			Ko:  info.KoreanName,
			En:  info.EnglishName,
			Ja:  info.JapaneseName,
		})
		if err != nil {
			panic(err)
		}
	}
}

// CompareBuddy compares buddy info from wiki with database
func (s *Service) CompareBuddy(info types.BuddyInfoFromAEWiki, id int) {
	ctx := context.Background()
	buddyID := fmt.Sprintf("b%d", info.GameID)

	fmt.Println()
	fmt.Printf("=== Comparing Buddy: %s ===\n", buddyID)
	fmt.Println()

	// Get buddy from database
	dbBuddy, err := s.queries.GetBuddyWithDetails(ctx, buddyID)
	if err != nil {
		fmt.Printf("Buddy %s not found in database\n", buddyID)
		return
	}

	// Get partner character ID if exists
	calculatedPartnerID := pgtype.Text{Valid: false}
	if info.PartnerLink != "None" && info.PartnerLink != "" {
		partnerID, err := s.queries.GetCharacterIDByWikiURL(ctx, pgtype.Text{String: info.PartnerLink, Valid: true})
		if err == nil {
			calculatedPartnerID = pgtype.Text{String: partnerID, Valid: true}
		}
	}

	// Compare all fields and display in color
	compareField := func(name string, dbVal, wikiVal interface{}) {
		dbStr := fmt.Sprintf("%v", dbVal)
		wikiStr := fmt.Sprintf("%v", wikiVal)

		if dbStr == wikiStr {
			// Same - Green
			fmt.Printf("  %-20s: %s%-30s%s (DB) = %s%-30s%s (Wiki)\n",
				name, common.ColorGreen, dbStr, common.ColorReset, common.ColorGreen, wikiStr, common.ColorReset)
		} else {
			// Different - Red
			fmt.Printf("  %-20s: %s%-30s%s (DB) ≠ %s%-30s%s (Wiki)\n",
				name, common.ColorRed, dbStr, common.ColorReset, common.ColorRed, wikiStr, common.ColorReset)
		}
	}

	compareField("EnglishName", dbBuddy.EnglishName.String, info.EnglishName)
	compareField("Style", dbBuddy.Style, string(info.Style))

	dbPartner := "NULL"
	if dbBuddy.PartnerAewikiUrl.Valid {
		dbPartner = dbBuddy.PartnerAewikiUrl.String
	}
	wikiPartner := "NULL"
	if calculatedPartnerID.Valid {
		wikiPartner = calculatedPartnerID.String
	}
	compareField("PartnerID", dbPartner, wikiPartner)

	compareField("WikiURL", dbBuddy.AewikiUrl.String, info.WikiURL)

	fmt.Println()
}

// UpsertBuddy inserts or updates a buddy (dual storage: buddies table + character.buddy_data JSONB)
func (s *Service) UpsertBuddy(info types.BuddyInfoFromAEWiki, dryrun bool) {
	if dryrun {
		fmt.Printf("[DRYRUN] UpsertBuddy: %s (id=%d)\n", info.EnglishName, info.GameID)
		return
	}

	ctx := context.Background()
	buddyID := fmt.Sprintf("b%d", info.GameID)

	// Get partner character ID if exists
	calculatedPartnerID := pgtype.Text{Valid: false}
	if info.PartnerLink != "None" && info.PartnerLink != "" {
		partnerID, err := s.queries.GetCharacterIDByWikiURL(ctx, pgtype.Text{String: info.PartnerLink, Valid: true})
		if err == nil {
			calculatedPartnerID = pgtype.Text{String: partnerID, Valid: true}
		}
	}

	// Check if buddy exists
	exists, err := s.queries.CheckBuddyExists(ctx, buddyID)
	if err != nil {
		panic(err)
	}

	// Store in buddies table
	if exists {
		fmt.Println("Buddy already exists: ", info.GameID)
		if calculatedPartnerID.Valid {
			err = s.queries.UpdateBuddyWithCharacter(ctx, postgres.UpdateBuddyWithCharacterParams{
				BuddyID:     buddyID,
				CharacterID: calculatedPartnerID,
				AewikiUrl:   pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl:   pgtype.Text{Valid: false},
			})
		} else {
			err = s.queries.UpdateBuddyWithGetPath(ctx, postgres.UpdateBuddyWithGetPathParams{
				BuddyID:   buddyID,
				GetPath:   pgtype.Text{String: "Unknown", Valid: true},
				AewikiUrl: pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl: pgtype.Text{Valid: false},
			})
		}
		if err != nil {
			panic(err)
		}
	} else {
		if calculatedPartnerID.Valid {
			err = s.queries.InsertBuddyWithCharacter(ctx, postgres.InsertBuddyWithCharacterParams{
				BuddyID:     buddyID,
				CharacterID: calculatedPartnerID,
				AewikiUrl:   pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl:   pgtype.Text{Valid: false},
			})
		} else {
			err = s.queries.InsertBuddyWithGetPath(ctx, postgres.InsertBuddyWithGetPathParams{
				BuddyID:   buddyID,
				GetPath:   pgtype.Text{String: "Unknown", Valid: true},
				AewikiUrl: pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl: pgtype.Text{Valid: false},
			})
		}
		if err != nil {
			panic(err)
		}
	}

	// Also store in character.buddy_data JSONB if partner character exists
	if calculatedPartnerID.Valid {
		// TODO: Store buddy info in character's buddy_data JSONB field
		// This will be implemented when we modify UpsertCharacter
	}
}
