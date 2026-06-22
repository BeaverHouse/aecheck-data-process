package batch

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	data "aecheck-data-process/internal/logic/external"
	"aecheck-data-process/internal/logic/parse"
	"aecheck-data-process/internal/types"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/BeaverHouse/go-common/logger"
)

type BuddyContext struct {
	Info        *types.BuddyInfoFromAEWiki
	BuddyID     string
	Translation *types.TranslationInfo
	SeesaaURL   string
}

func findBuddyTranslation(buddyID string, englishName string, dbService *database.Service) *types.TranslationInfo {
	row, err := dbService.GetTranslation(context.Background(), buddyID)
	if err == nil {
		return &types.TranslationInfo{
			EnglishName:  row.En,
			KoreanName:   row.Ko,
			JapaneseName: row.Ja,
		}
	}
	common.Log.Warn("No DB translation for buddy", logger.Field{Key: "name", Value: englishName}, logger.Field{Key: "buddyID", Value: buddyID})
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("  Korean  : ")
	ko, _ := reader.ReadString('\n')
	fmt.Print("  Japanese: ")
	ja, _ := reader.ReadString('\n')
	return &types.TranslationInfo{
		EnglishName:  englishName,
		KoreanName:   strings.TrimSpace(ko),
		JapaneseName: strings.TrimSpace(ja),
	}
}

func ResolveBuddy(wikiURL string, dbService *database.Service) *BuddyContext {
	doc, err := data.GetDocumentFromURL(wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to load document: %v", err))
	}

	info, err := parse.GetBuddyInfoFromAEWiki(doc, wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to get buddy info: %v", err))
	}
	buddyID := dbService.ResolveBuddyID(wikiURL)
	translation := findBuddyTranslation(buddyID, info.EnglishName, dbService)
	seesaaURL := logic.FindBuddySeesaaLink(info.Style, translation.JapaneseName)
	return &BuddyContext{
		Info:        info,
		BuddyID:     buddyID,
		Translation: translation,
		SeesaaURL:   seesaaURL,
	}
}

func CompareBuddy(ctx *BuddyContext, dbService *database.Service) {
	dbService.CompareBuddy(*ctx.Info, ctx.BuddyID, ctx.SeesaaURL)
	dbService.CompareTranslations(*ctx.Translation, ctx.BuddyID)
}

func UpdateBuddy(ctx *BuddyContext, dryrun bool, dbService *database.Service) {
	dbService.UpsertBuddy(*ctx.Info, ctx.BuddyID, ctx.SeesaaURL, dryrun)
	if err := data.UploadBuddyImage(ctx.Info.GameID, ctx.BuddyID, dryrun); err != nil {
		panic(fmt.Sprintf("Failed to upload buddy image: %v", err))
	}
	dbService.UpsertTranslation(*ctx.Translation, ctx.BuddyID, dryrun)
}
