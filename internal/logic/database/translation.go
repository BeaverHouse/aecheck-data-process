package database

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/types"
	"context"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
)

// CompareTranslations compares translation info with database
func (s *Service) CompareTranslations(info types.TranslationInfo, code string) {
	ctx := context.Background()

	common.Log.Info("Translation comparison", logger.Field{Key: "code", Value: code})

	trans, err := s.queries.GetTranslation(ctx, code)
	if err != nil {
		common.Log.Warn("Translation not found in database")
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

	dbEn := logic.DisplayEnglishName(trans.En)
	matched := logic.EnglishNamesMatch(trans.En, info.EnglishName)
	color := common.ColorGreen
	sym := "="
	if !matched {
		color = common.ColorRed
		sym = "≠"
	}
	fmt.Printf("  %-12s: %s%-20s%s (DB) %s %s%-20s%s (Wiki)\n",
		"English", color, dbEn, common.ColorReset, sym, color, info.EnglishName, common.ColorReset)

	compareField("Japanese", trans.Ja, info.JapaneseName)
	fmt.Println()
}

// UpsertTranslation inserts or updates translation
func (s *Service) UpsertTranslation(info types.TranslationInfo, code string, dryrun bool) {
	if logic.IsSpoilerTrueName(info.EnglishName) {
		info.EnglishName = logic.ResolveForDB(info.EnglishName)
	}

	if dryrun {
		common.Log.Info("[DRYRUN] UpsertTranslation", logger.Field{Key: "code", Value: code}, logger.Field{Key: "en", Value: info.EnglishName})
		return
	}

	ctx := context.Background()

	exists, err := s.queries.CheckTranslationExists(ctx, code)
	if err != nil {
		panic(err)
	}

	if exists {
		common.Log.Info("Translation already exists", logger.Field{Key: "code", Value: code})
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
