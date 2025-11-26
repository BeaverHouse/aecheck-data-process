package batch

import (
	"aecheck-data-process/internal/constants"
	"aecheck-data-process/internal/types"
	"encoding/csv"
	"os"
	"strconv"
)

func getID(englishWord string, englishClassName string) int {
	file, err := os.Open("internal/logic/batch/files/i18n.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	if constants.HIDDEN_NAMES[englishWord] != "" {
		englishWord = constants.HIDDEN_NAMES[englishWord]
	}

	for _, record := range records {
		if record[1] == englishWord && record[2] == englishClassName {
			id, err := strconv.Atoi(record[0])
			if err != nil {
				panic(err)
			}
			return id
		}
	}

	return -1
}

func findTranslation(englishWord string, englishClassName string, isClass bool) *types.TranslationInfo {
	file, err := os.Open("internal/logic/batch/files/i18n.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	// Skip header
	_, err = reader.Read()
	if err != nil {
		panic(err)
	}

	records, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}

	if constants.HIDDEN_NAMES[englishWord] != "" {
		englishWord = constants.HIDDEN_NAMES[englishWord]
	}

	for _, record := range records {
		if len(record) < 6 {
			continue
		}

		if isClass {
			if record[1] == englishWord && record[2] == englishClassName {
				return &types.TranslationInfo{
					EnglishName:  record[2],
					KoreanName:   record[4],
					JapaneseName: record[6],
				}
			}
		} else {
			if record[1] == englishWord && record[2] == englishClassName {
				return &types.TranslationInfo{
					EnglishName:  record[1],
					KoreanName:   record[3],
					JapaneseName: record[5],
				}
			}
		}

	}

	return nil
}
