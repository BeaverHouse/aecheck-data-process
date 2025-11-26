package logic

import (
	"aecheck-data-process/internal/types"
	"strings"
)

func CheckAwaken(rawString string) bool {
	return strings.Contains(rawString, "Stellar Awakened")
}

func ClassifyLightShadow(rawString string) types.AELightShadow {
	if strings.Contains(strings.ToLower(rawString), "light") {
		return types.LSLight
	}
	return types.LSShadow
}

func ClassifyCategory(rawString string) types.AECategory {
	if strings.Contains(rawString, "Dreams") {
		return types.AECategoryEncounter
	} else if strings.Contains(rawString, "Symphony") {
		return types.AECategoryColab
	}
	return types.AECategoryFree
}
