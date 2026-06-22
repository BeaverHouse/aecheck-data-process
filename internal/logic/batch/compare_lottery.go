package batch

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	data "aecheck-data-process/internal/logic/external"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/PuerkitoBio/goquery"
)

const (
	lotteryBaseURL = "https://api-us.another-eden.games/asset/lottery_notice/view/"
)

// LotteryTranslation represents a translation entry from the lottery API
type LotteryTranslation struct {
	Japanese string
	Korean   string
	English  string
}

// TranslationDiff represents a difference between API and DB
type TranslationDiff struct {
	Key      string
	Field    string // "ko" or "en"
	DBValue  string
	APIValue string
	Japanese string
}

// CompareLotteryTranslations fetches lottery data and compares with DB translations
func CompareLotteryTranslations(lotteryID string, dbService *database.Service) {
	// 1. Fetch data for all 3 languages
	jaData, err := fetchLotteryFirstColumn(lotteryID, "ja")
	if err != nil {
		common.Log.Error("Failed to fetch Japanese lottery data", logger.Field{Key: "error", Value: err})
		return
	}

	koData, err := fetchLotteryFirstColumn(lotteryID, "ko")
	if err != nil {
		common.Log.Error("Failed to fetch Korean lottery data", logger.Field{Key: "error", Value: err})
		return
	}

	enData, err := fetchLotteryFirstColumn(lotteryID, "en")
	if err != nil {
		common.Log.Error("Failed to fetch English lottery data", logger.Field{Key: "error", Value: err})
		return
	}

	// 2. Build translation map (ja -> {ko, en})
	translations := buildTranslationMap(jaData, koData, enData)

	// 3. Compare with DB and collect differences
	diffs := compareWithDB(translations, dbService)

	// 4. Display differences
	if len(diffs) == 0 {
		common.Log.Info("No differences found between API and DB translations")
		return
	}

	common.Log.Info("Found differences", logger.Field{Key: "count", Value: len(diffs)})
	for i, diff := range diffs {
		fmt.Printf("%d. [%s] %s (JA: %s)\n", i+1, diff.Key, diff.Field, diff.Japanese)
		fmt.Printf("   DB:  %s%s%s\n", common.ColorRed, diff.DBValue, common.ColorReset)
		fmt.Printf("   API: %s%s%s\n", common.ColorGreen, diff.APIValue, common.ColorReset)
		fmt.Println()
	}

	// 5. Ask user for confirmation
	fmt.Print("Apply these changes? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "y" || input == "yes" {
		applyDiffs(diffs, dbService)
		common.Log.Info("Changes applied successfully")
	} else {
		common.Log.Info("Changes not applied")
	}
}

// fetchLotteryFirstColumn fetches the lottery page and extracts the first column data
func fetchLotteryFirstColumn(lotteryID, language string) ([]string, error) {
	url := fmt.Sprintf("%s%s?language=%s", lotteryBaseURL, lotteryID, language)

	doc, err := data.GetDocumentFromURL(url)
	if err != nil {
		return nil, err
	}

	var results []string

	// Find the table and extract first column (skip header)
	doc.Find("table tr").Each(func(i int, s *goquery.Selection) {
		// Skip header row
		if i == 0 {
			return
		}

		// Get first td
		firstTd := s.Find("td").First()
		text := strings.TrimSpace(firstTd.Text())
		if text != "" {
			results = append(results, text)
		}
	})

	return results, nil
}

// buildTranslationMap creates a map from Japanese text to all translations
func buildTranslationMap(jaData, koData, enData []string) map[string]LotteryTranslation {
	translations := make(map[string]LotteryTranslation)

	// All arrays should have the same length
	minLen := len(jaData)
	if len(koData) < minLen {
		minLen = len(koData)
	}
	if len(enData) < minLen {
		minLen = len(enData)
	}

	for i := 0; i < minLen; i++ {
		ja := jaData[i]
		translations[ja] = LotteryTranslation{
			Japanese: ja,
			Korean:   koData[i],
			English:  enData[i],
		}
	}

	return translations
}

// isASCII checks if a string contains only ASCII characters
func isASCII(s string) bool {
	for _, r := range s {
		if r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// compareWithDB compares API translations with DB and returns differences
func compareWithDB(translations map[string]LotteryTranslation, dbService *database.Service) []TranslationDiff {
	ctx := context.Background()

	jas := make([]string, 0, len(translations))
	for ja := range translations {
		jas = append(jas, ja)
	}

	dbRows, err := dbService.GetTranslationsByJapaneseList(ctx, jas)
	if err != nil {
		common.Log.Error("Failed to fetch DB translations", logger.Field{Key: "error", Value: err})
		return nil
	}

	dbByJa := make(map[string]postgres.GetTranslationsByJapaneseListRow, len(dbRows))
	for _, row := range dbRows {
		dbByJa[row.Ja] = row
	}

	var diffs []TranslationDiff
	for ja, apiTrans := range translations {
		dbTrans, ok := dbByJa[ja]
		if !ok {
			continue
		}

		if dbTrans.Ko != apiTrans.Korean {
			diffs = append(diffs, TranslationDiff{
				Key:      dbTrans.Key,
				Field:    "ko",
				DBValue:  dbTrans.Ko,
				APIValue: apiTrans.Korean,
				Japanese: ja,
			})
		}

		if dbTrans.En != apiTrans.English {
			if isASCII(apiTrans.English) {
				diffs = append(diffs, TranslationDiff{
					Key:      dbTrans.Key,
					Field:    "en",
					DBValue:  dbTrans.En,
					APIValue: apiTrans.English,
					Japanese: ja,
				})
			} else {
				common.Log.Warn("Skipped non-ASCII translation", logger.Field{Key: "key", Value: dbTrans.Key}, logger.Field{Key: "dbEn", Value: dbTrans.En}, logger.Field{Key: "apiEn", Value: apiTrans.English})
			}
		}
	}

	return diffs
}

// applyDiffs applies the differences to the database
func applyDiffs(diffs []TranslationDiff, dbService *database.Service) {
	ctx := context.Background()

	// Group diffs by key to update each row once
	keyUpdates := make(map[string]struct {
		Ko *string
		En *string
	})

	for _, diff := range diffs {
		update := keyUpdates[diff.Key]
		if diff.Field == "ko" {
			update.Ko = &diff.APIValue
		} else if diff.Field == "en" {
			update.En = &diff.APIValue
		}
		keyUpdates[diff.Key] = update
	}

	for key, update := range keyUpdates {
		// Get current values
		current, err := dbService.GetTranslation(ctx, key)
		if err != nil {
			common.Log.Error("Error getting translation", logger.Field{Key: "key", Value: key}, logger.Field{Key: "error", Value: err})
			continue
		}

		// Apply updates
		newKo := current.Ko
		newEn := current.En
		newJa := current.Ja

		if update.Ko != nil {
			newKo = *update.Ko
		}
		if update.En != nil {
			newEn = *update.En
		}

		err = dbService.UpdateTranslationValues(ctx, key, newKo, newEn, newJa)
		if err != nil {
			common.Log.Error("Error updating translation", logger.Field{Key: "key", Value: key}, logger.Field{Key: "error", Value: err})
		}
	}
}
