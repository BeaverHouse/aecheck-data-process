package batch

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	data "aecheck-data-process/internal/logic/external"
	"aecheck-data-process/internal/logic/parse"
	"aecheck-data-process/internal/types"
	"fmt"

	"github.com/BeaverHouse/go-common/logger"
)

type CharacterContext struct {
	info             *types.CharacterInfoFromAEWiki
	nameTranslation  *types.TranslationInfo
	classTranslation *types.TranslationInfo
	id               int
	seesaaURL        string
}

func ResolveCharacter(wikiURL string, dbService *database.Service) *CharacterContext {
	doc, err := data.GetDocumentFromURL(wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to load document: %v", err))
	}

	info, err := parse.GetCharacterInfo(doc, wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to get character info: %v", err))
	}
	nameTr := findTranslation(info.EnglishName, info.EnglishClassName, false, dbService)
	classTr := findTranslation(info.EnglishName, info.EnglishClassName, true, dbService)
	id := getID(info.GameID, string(info.Style), info.EnglishName, info.EnglishClassName, dbService)
	seesaaURL := logic.FindSeesaaLink(*info, nameTr.JapaneseName)

	return &CharacterContext{
		info:             info,
		nameTranslation:  nameTr,
		classTranslation: classTr,
		id:               id,
		seesaaURL:        seesaaURL,
	}
}

func CompareCharacter(ctx *CharacterContext, dbService *database.Service) {
	dbService.CompareCharacter(*ctx.info, ctx.seesaaURL, ctx.id)
	dbService.CompareDungeon(ctx.id, *ctx.info)
	dbService.ComparePersonality(ctx.id, *ctx.info)
	dbService.CompareTranslations(*ctx.nameTranslation, fmt.Sprintf("c%d", ctx.info.GameID))
	dbService.CompareTranslations(*ctx.classTranslation, fmt.Sprintf("book.char%04d", ctx.id))
}

func UpdateCharacter(ctx *CharacterContext, dryrun bool, dbService *database.Service) {
	info := ctx.info
	id := ctx.id

	characterID, fourStarStatus, isFirstFiveStar := dbService.CheckFourStarUpdate(*info, false)
	common.Log.Info("4-star status", logger.Field{Key: "status", Value: fourStarStatus}, logger.Field{Key: "characterID", Value: characterID}, logger.Field{Key: "isFirstFiveStar", Value: isFirstFiveStar})
	switch fourStarStatus {
	case types.NotExists:
		dbService.UpsertCharacter(*info, ctx.seesaaURL, id-1, dryrun)
		dbService.InsertFourStarCharacter(id-1, dryrun)
		if err := data.UploadCharacterImage(*info, id-1, true, dryrun); err != nil {
			panic(fmt.Sprintf("Failed to upload 4-star character image: %v", err))
		}
	case types.NotUpdated:
		dbService.UpsertCharacter(*info, ctx.seesaaURL, characterID, dryrun)
		dbService.UpdateFourStarCharacter(characterID, dryrun)
		dbService.UpsertCharacter(*info, ctx.seesaaURL, id, dryrun)
	default:
		dbService.UpsertCharacter(*info, ctx.seesaaURL, id, dryrun)
	}
	if err := data.UploadCharacterImage(*info, id, false, dryrun); err != nil {
		panic(fmt.Sprintf("Failed to upload character image: %v", err))
	}
	dbService.UpsertTranslation(*ctx.nameTranslation, fmt.Sprintf("c%d", info.GameID), dryrun)
	dbService.UpsertTranslation(*ctx.classTranslation, fmt.Sprintf("book.char%04d", id), dryrun)
}
