package batch

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	"aecheck-data-process/internal/types"
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/BeaverHouse/go-common/logger"
)

func getID(gameID int, style string, englishName string, englishClassName string, dbService *database.Service) int {
	id, err := dbService.GetCharacterNumericID(context.Background(), gameID, style)
	if err != nil {
		return promptID(englishName, englishClassName)
	}
	return id
}

func findTranslation(englishName string, englishClassName string, isClass bool, dbService *database.Service) *types.TranslationInfo {
	target := englishName
	if isClass {
		target = englishClassName
	}
	_, info, err := dbService.FindTranslationFromDB(context.Background(), target, isClass)
	if err == nil {
		return info
	}
	return promptTranslation(englishName, englishClassName, isClass)
}

func findSpoilerTranslation(englishName string, code string, dbService *database.Service) *types.TranslationInfo {
	if englishName == "" {
		return nil
	}
	if !logic.IsSpoilerTrueName(englishName) {
		return nil
	}

	row, err := dbService.GetTranslation(context.Background(), "spoiler."+code)
	if err == nil && row.En == englishName {
		return &types.TranslationInfo{
			EnglishName:  row.En,
			KoreanName:   row.Ko,
			JapaneseName: row.Ja,
		}
	}
	return promptSpoilerTranslation(englishName)
}

func promptID(englishName, englishClassName string) int {
	common.Log.Warn("No DB entry found", logger.Field{Key: "name", Value: englishName}, logger.Field{Key: "class", Value: englishClassName})
	fmt.Print("Enter numeric ID: ")
	var input string
	fmt.Fscan(os.Stdin, &input)
	id, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		panic(fmt.Sprintf("invalid ID input: %s", input))
	}
	return id
}

func promptTranslation(englishName, englishClassName string, isClass bool) *types.TranslationInfo {
	label := englishName
	if isClass {
		label = englishClassName
	}
	common.Log.Warn("No DB translation found", logger.Field{Key: "name", Value: label}, logger.Field{Key: "isClass", Value: isClass})
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("  Korean  : ")
	ko, _ := reader.ReadString('\n')
	fmt.Print("  Japanese: ")
	ja, _ := reader.ReadString('\n')

	en := label
	return &types.TranslationInfo{
		EnglishName:  en,
		KoreanName:   strings.TrimSpace(ko),
		JapaneseName: strings.TrimSpace(ja),
	}
}

func promptSpoilerTranslation(englishName string) *types.TranslationInfo {
	common.Log.Warn("Enter spoiler true-name translation", logger.Field{Key: "name", Value: englishName})
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
