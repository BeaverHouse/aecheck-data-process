package batch

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/data"
	"aecheck-data-process/internal/logic/database"
	"aecheck-data-process/internal/logic/parse"
	"aecheck-data-process/internal/types"
	"fmt"
)

func CompareCharacter(wikiURL string, dbService *database.Service) {
	info, err := parse.GetCharacterInfo(wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to get character info: %v", err))
	}
	nameTranslation := findTranslation(info.EnglishName, info.EnglishClassName, false)
	classTranslation := findTranslation(info.EnglishName, info.EnglishClassName, true)
	id := getID(info.EnglishName, info.EnglishClassName)

	scrapedSeesaaURL := logic.FindSeesaaLink(*info, nameTranslation.JapaneseName)

	dbService.CompareCharacter(*info, scrapedSeesaaURL, id)
	dbService.CompareDungeon(id, *info)
	dbService.ComparePersonality(id, *info)
	dbService.CompareTranslations(*nameTranslation, fmt.Sprintf("c%d", info.GameID))
	dbService.CompareTranslations(*classTranslation, fmt.Sprintf("book.char%04d", id))
}

func UpdateCharacter(wikiURL string, dryrun bool, dbService *database.Service) {
	info, err := parse.GetCharacterInfo(wikiURL)
	if err != nil {
		panic(fmt.Sprintf("Failed to get character info: %v", err))
	}
	nameTranslation := findTranslation(info.EnglishName, info.EnglishClassName, false)
	classTranslation := findTranslation(info.EnglishName, info.EnglishClassName, true)
	id := getID(info.EnglishName, info.EnglishClassName)

	scrapedSeesaaURL := logic.FindSeesaaLink(*info, nameTranslation.JapaneseName)

	characterID, fourStarStatus := dbService.CheckFourStarUpdate(*info, false)
	fmt.Println("4-star status: ", fourStarStatus, "characterID: ", characterID)
	switch fourStarStatus {
	case types.NotExists:
		dbService.UpsertCharacter(*info, scrapedSeesaaURL, id-1, dryrun)
		dbService.InsertFourStarCharacter(id-1, dryrun)
		data.UploadCharacterImage(*info, id-1, true, dryrun)
		dbService.UpsertDungeon(id-1, *info, dryrun)
		dbService.UpsertPersonality(id-1, *info, dryrun)
	case types.NotUpdated:
		dbService.UpsertCharacter(*info, scrapedSeesaaURL, characterID, dryrun)
		dbService.UpdateFourStarCharacter(characterID, dryrun)
		dbService.UpsertDungeon(characterID, *info, dryrun)
		dbService.UpsertPersonality(characterID, *info, dryrun)
		dbService.PurgeDeletedDungeon(characterID, dryrun)
		dbService.PurgeDeletedPersonality(characterID, dryrun)
		dbService.UpsertCharacter(*info, scrapedSeesaaURL, id, dryrun)
	default:
		dbService.UpsertCharacter(*info, scrapedSeesaaURL, id, dryrun)
	}
	data.UploadCharacterImage(*info, id, false, dryrun)
	dbService.UpsertDungeon(id, *info, dryrun)
	dbService.UpsertPersonality(id, *info, dryrun)
	dbService.UpsertTranslation(*nameTranslation, fmt.Sprintf("c%d", info.GameID), dryrun)
	dbService.UpsertTranslation(*classTranslation, fmt.Sprintf("book.char%04d", id), dryrun)

	fmt.Println("Purging deleted dungeons and personalities")
	dbService.PurgeDeletedDungeon(id, dryrun)
	dbService.PurgeDeletedPersonality(id, dryrun)
}
