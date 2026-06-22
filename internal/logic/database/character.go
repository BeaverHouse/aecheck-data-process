package database

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

// resolveAlterCharacterID resolves the alter character ID from the wiki link, returning an invalid pgtype.Text if not found.
func (s *Service) resolveAlterCharacterID(ctx context.Context, info types.CharacterInfoFromAEWiki) pgtype.Text {
	link := logic.FindAlterLink(info)
	if link == "" {
		return pgtype.Text{Valid: false}
	}
	alterCode, err := s.queries.GetCharacterCodeByWikiURL(ctx, pgtype.Text{String: link, Valid: true})
	if err != nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: alterCode, Valid: true}
}

// CompareCharacter compares character info from wiki with database
func (s *Service) CompareCharacter(info types.CharacterInfoFromAEWiki, scrapedSeesaaURL string, id int) {
	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	common.Log.Info("Comparing Character", logger.Field{Key: "characterID", Value: characterID})

	dbChar, err := s.queries.GetCharacterWithTranslation(ctx, characterID)
	if err != nil {
		common.Log.Warn("Character not found in database - New character from Wiki", logger.Field{Key: "characterID", Value: characterID})

		calculatedAlterCharacterID := s.resolveAlterCharacterID(ctx, info)

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

	calculatedAlterCharacterID := s.resolveAlterCharacterID(ctx, info)

	compareField := func(name string, dbVal, wikiVal interface{}) {
		dbStr := fmt.Sprintf("%v", dbVal)
		wikiStr := fmt.Sprintf("%v", wikiVal)

		if dbStr == wikiStr {
			fmt.Printf("  %-20s: %s%-30s%s (DB) = %s%-30s%s (Wiki)\n",
				name, common.ColorGreen, dbStr, common.ColorReset, common.ColorGreen, wikiStr, common.ColorReset)
		} else {
			fmt.Printf("  %-20s: %s%-30s%s (DB) ≠ %s%-30s%s (Wiki)\n",
				name, common.ColorRed, dbStr, common.ColorReset, common.ColorRed, wikiStr, common.ColorReset)
		}
	}

	dbEn := logic.DisplayEnglishName(dbChar.EnglishName.String)
	matched := logic.EnglishNamesMatch(dbChar.EnglishName.String, info.EnglishName)
	color := common.ColorGreen
	sym := "="
	if !matched {
		color = common.ColorRed
		sym = "≠"
	}
	fmt.Printf("  %-20s: %s%-30s%s (DB) %s %s%-30s%s (Wiki)\n",
		"EnglishName", color, dbEn, common.ColorReset, sym, color, info.EnglishName, common.ColorReset)

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
		common.Log.Info("[DRYRUN] UpsertCharacter", logger.Field{Key: "name", Value: info.EnglishName}, logger.Field{Key: "id", Value: id})
		return
	}

	ctx := context.Background()
	characterID := fmt.Sprintf("char%04d", id)

	calculatedAlterCharacterID := s.resolveAlterCharacterID(ctx, info)

	updateDate := pgtype.Date{Valid: false}
	if info.UpdateDate != "" {
		t, err := time.Parse("2006-01-02", info.UpdateDate)
		if err == nil {
			updateDate = pgtype.Date{Time: t, Valid: true}
		}
	}

	personalitiesData, err := s.convertPersonalitiesToJSONB(ctx, info.Personalities)
	if err != nil {
		panic(err)
	}

	dungeonsData, err := s.convertDungeonsToJSONB(ctx, info.Dungeons)
	if err != nil {
		panic(err)
	}

	var buddyData []byte = nil

	exists, err := s.queries.CheckCharacterExists(ctx, characterID)
	if err != nil {
		panic(err)
	}

	if exists {
		common.Log.Info("Character already exists", logger.Field{Key: "id", Value: id})
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

// CheckFourStarUpdate checks if a four-star character needs updating
func (s *Service) CheckFourStarUpdate(info types.CharacterInfoFromAEWiki, excluded bool) (int, types.UpdateStatus, bool) {
	if !info.HasFourStar {
		common.Log.Info("No 4-star rarity", logger.Field{Key: "name", Value: info.EnglishName})
		return -1, types.NotNeeded, false
	} else if info.IsAlter {
		common.Log.Info("Alter character starts with 5-star in default", logger.Field{Key: "name", Value: info.EnglishName})
		return -1, types.NotNeeded, false
	} else if info.Style == types.StyleES {
		common.Log.Info("ES character differs weapon type from default", logger.Field{Key: "name", Value: info.EnglishName})
		return -1, types.NotNeeded, false
	} else if excluded {
		common.Log.Info("Excluded character", logger.Field{Key: "name", Value: info.EnglishName})
		return -1, types.NotNeeded, false
	}

	ctx := context.Background()
	characterCode := fmt.Sprintf("c%d", info.GameID)

	fiveStarExists, err := s.queries.CheckFiveStarExistsByCode(ctx, characterCode)
	if err != nil {
		panic(err)
	}

	if fiveStarExists {
		common.Log.Info("5-star style already exists", logger.Field{Key: "name", Value: info.EnglishName})
		return -1, types.NotNeeded, false
	}

	char, err := s.queries.GetFirstCharacterByCode(ctx, characterCode)
	if err != nil {
		return -1, types.NotExists, true
	}

	id, err := strconv.Atoi(strings.Replace(char.CharacterID, "char", "", 1))
	if err != nil {
		panic(err)
	}

	if char.Style != "☆4" {
		return -1, types.NotNeeded, true
	} else if char.UpdatedAt.Valid && char.UpdatedAt.Time.After(time.Now().AddDate(0, -6, 0)) {
		return id, types.Updated, true
	} else {
		return id, types.NotUpdated, true
	}
}

// InsertFourStarCharacter inserts a four-star character by copying from existing ID
func (s *Service) InsertFourStarCharacter(id int, dryrun bool) {
	if dryrun {
		common.Log.Info("[DRYRUN] InsertFourStarCharacter", logger.Field{Key: "id", Value: id})
		return
	}

	ctx := context.Background()

	err := s.queries.CopyCharacterToNewID(ctx, postgres.CopyCharacterToNewIDParams{
		CharacterID:   fmt.Sprintf("char%04d", id+1),
		CharacterID_2: fmt.Sprintf("char%04d", id),
	})
	if err != nil {
		panic(err)
	}

	err = s.queries.UpdateCharacterToFourStar(ctx, fmt.Sprintf("char%04d", id))
	if err != nil {
		panic(err)
	}
}

// UpdateFourStarCharacter updates a character to four-star
func (s *Service) UpdateFourStarCharacter(id int, dryrun bool) {
	if dryrun {
		common.Log.Info("[DRYRUN] UpdateFourStarCharacter", logger.Field{Key: "id", Value: id})
		return
	}

	ctx := context.Background()
	err := s.queries.UpdateCharacterToFourStar(ctx, fmt.Sprintf("char%04d", id))
	if err != nil {
		panic(err)
	}
}
