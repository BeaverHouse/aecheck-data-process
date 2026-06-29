package logic

import (
	"aecheck-data-process/internal/constants"
	"fmt"
)

// ResolveSpoilerName returns the true name if given a spoiler alias, otherwise returns name as-is.
// e.g., "Red Clad Flam." → "Dewey (Alter)", "Vares" → "Vares"
func ResolveSpoilerName(name string) string {
	if trueName, ok := constants.SPOILER_NAMES[name]; ok {
		return trueName
	}
	return name
}

// IsSpoilerTrueName returns true if name is a spoiler true name (value in SPOILER_NAMES).
// Use this to skip DB writes where the alias is intentionally stored instead of the true name.
// e.g., DB stores "Red Clad Flam." — writing "Dewey (Alter)" would overwrite the spoiler-safe alias.
func IsSpoilerTrueName(name string) bool {
	for _, trueName := range constants.SPOILER_NAMES {
		if trueName == name {
			return true
		}
	}
	return false
}

// IsSpoilerAlias returns true if name is a spoiler-safe alias stored in DB.
func IsSpoilerAlias(name string) bool {
	_, ok := constants.SPOILER_NAMES[name]
	return ok
}

// ResolveForDB returns the name to use when searching the DB.
// All SPOILER_NAMES characters have their alias stored in DB (not the true name),
// so both alias→trueName and trueName→alias lookups resolve to the DB-stored alias.
func ResolveForDB(name string) string {
	if _, ok := constants.SPOILER_NAMES[name]; ok {
		return name // already the alias
	}
	for alias, trueName := range constants.SPOILER_NAMES {
		if trueName == name {
			return alias
		}
	}
	return name
}

// DisplayEnglishName formats a DB english name for display.
// If it's a spoiler alias, shows "alias (aka trueName)".
func DisplayEnglishName(dbName string) string {
	if trueName, ok := constants.SPOILER_NAMES[dbName]; ok {
		return fmt.Sprintf("%s (aka %s)", dbName, trueName)
	}
	return dbName
}

// EnglishNamesMatch returns true if dbName and wikiName refer to the same character,
// accounting for spoiler alias/true-name equivalence.
func EnglishNamesMatch(dbName, wikiName string) bool {
	if dbName == wikiName {
		return true
	}
	if trueName, ok := constants.SPOILER_NAMES[dbName]; ok && trueName == wikiName {
		return true
	}
	return false
}
