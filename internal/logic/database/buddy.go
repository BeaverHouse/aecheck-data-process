package database

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"context"
	"encoding/json"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

// resolvePartnerID resolves the partner character ID from the wiki link, returning an invalid pgtype.Text if not found.
func (s *Service) resolvePartnerID(ctx context.Context, partnerLink string) pgtype.Text {
	if partnerLink == "" || partnerLink == "None" {
		return pgtype.Text{Valid: false}
	}
	partnerID, err := s.queries.GetCharacterIDByWikiURL(ctx, pgtype.Text{String: partnerLink, Valid: true})
	if err != nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: partnerID, Valid: true}
}

// CompareBuddy compares buddy info from wiki with database
func (s *Service) CompareBuddy(info types.BuddyInfoFromAEWiki, buddyID string, seesaaURL string) {
	ctx := context.Background()

	common.Log.Info("Comparing Buddy", logger.Field{Key: "buddyID", Value: buddyID})

	dbBuddy, err := s.queries.GetBuddyWithDetails(ctx, buddyID)
	if err != nil {
		common.Log.Warn("Buddy not found in database - New buddy from Wiki", logger.Field{Key: "buddyID", Value: buddyID})
		fmt.Printf("  %-20s: %s\n", "EnglishName", info.EnglishName)
		fmt.Printf("  %-20s: %s\n", "Style", info.Style)
		fmt.Printf("  %-20s: %s\n", "PartnerLink", info.PartnerLink)
		fmt.Printf("  %-20s: %s\n", "WikiURL", info.WikiURL)
		fmt.Printf("  %-20s: %s\n", "SeesaaURL", seesaaURL)
		fmt.Println()
		return
	}

	calculatedPartnerID := s.resolvePartnerID(ctx, info.PartnerLink)

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

	compareField("EnglishName", dbBuddy.EnglishName.String, info.EnglishName)

	dbPartner := "NULL"
	if dbBuddy.PartnerCharacterID.Valid {
		dbPartner = dbBuddy.PartnerCharacterID.String
	}
	wikiPartner := "NULL"
	if calculatedPartnerID.Valid {
		wikiPartner = calculatedPartnerID.String
	}
	compareField("PartnerID", dbPartner, wikiPartner)

	compareField("WikiURL", dbBuddy.AewikiUrl.String, info.WikiURL)
	compareField("SeesaaURL", dbBuddy.SeesaaUrl.String, seesaaURL)

	fmt.Println()
}

// ResolveBuddyID returns existing buddy_id by wiki URL, or generates a new one.
func (s *Service) ResolveBuddyID(wikiURL string) string {
	ctx := context.Background()
	existingID, err := s.queries.GetBuddyIDByWikiURL(ctx, pgtype.Text{String: wikiURL, Valid: true})
	if err == nil {
		return existingID
	}
	count, err := s.queries.GetBuddyCount(ctx)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("buddy%04d", 2000+count+1)
}

// UpsertBuddy inserts or updates a buddy and updates partner character's buddy_data
func (s *Service) UpsertBuddy(info types.BuddyInfoFromAEWiki, buddyID string, seesaaURL string, dryrun bool) {
	if dryrun {
		common.Log.Info("[DRYRUN] UpsertBuddy", logger.Field{Key: "name", Value: info.EnglishName}, logger.Field{Key: "buddyID", Value: buddyID})
		fmt.Printf("  %-20s: %s\n", "BuddyID", buddyID)
		fmt.Printf("  %-20s: %s\n", "Style", info.Style)
		fmt.Printf("  %-20s: %s\n", "PartnerLink", info.PartnerLink)
		fmt.Printf("  %-20s: %s\n", "WikiURL", info.WikiURL)
		fmt.Printf("  %-20s: %s\n", "SeesaaURL", seesaaURL)
		return
	}

	ctx := context.Background()

	calculatedPartnerID := s.resolvePartnerID(ctx, info.PartnerLink)

	exists, err := s.queries.CheckBuddyExists(ctx, buddyID)
	if err != nil {
		panic(err)
	}

	seesaaText := pgtype.Text{String: seesaaURL, Valid: seesaaURL != ""}

	if exists {
		common.Log.Info("Buddy already exists", logger.Field{Key: "buddyID", Value: buddyID})
		if calculatedPartnerID.Valid {
			err = s.queries.UpdateBuddyWithCharacter(ctx, postgres.UpdateBuddyWithCharacterParams{
				BuddyID:     buddyID,
				CharacterID: calculatedPartnerID,
				AewikiUrl:   pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl:   seesaaText,
			})
		} else {
			err = s.queries.UpdateBuddyWithGetPath(ctx, postgres.UpdateBuddyWithGetPathParams{
				BuddyID:   buddyID,
				GetPath:   pgtype.Text{String: "Unknown", Valid: true},
				AewikiUrl: pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl: seesaaText,
			})
		}
	} else {
		if calculatedPartnerID.Valid {
			err = s.queries.InsertBuddyWithCharacter(ctx, postgres.InsertBuddyWithCharacterParams{
				BuddyID:     buddyID,
				CharacterID: calculatedPartnerID,
				AewikiUrl:   pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl:   seesaaText,
			})
		} else {
			err = s.queries.InsertBuddyWithGetPath(ctx, postgres.InsertBuddyWithGetPathParams{
				BuddyID:   buddyID,
				GetPath:   pgtype.Text{String: "Unknown", Valid: true},
				AewikiUrl: pgtype.Text{String: info.WikiURL, Valid: true},
				SeesaaUrl: seesaaText,
			})
		}
	}
	if err != nil {
		panic(err)
	}

	if calculatedPartnerID.Valid {
		buddyData, err := json.Marshal(map[string]string{"buddy_id": buddyID})
		if err != nil {
			panic(err)
		}
		err = s.queries.UpdateCharacterBuddyData(ctx, postgres.UpdateCharacterBuddyDataParams{
			CharacterID: calculatedPartnerID.String,
			BuddyData:   buddyData,
		})
		if err != nil {
			panic(err)
		}
		common.Log.Info("Updated buddy_data", logger.Field{Key: "partnerID", Value: calculatedPartnerID.String})
	}
}
