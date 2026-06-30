package constants

const (
	AEWIKI_BASE_URL = "https://anothereden.wiki/w/"
	SEESAA_BASE_URL = "https://anothereden.game-info.wiki/d/"
)

// Style suffix for character image filename.
// e.g., NS -> "_rank5", AS -> "_s2_rank5", ES -> "_s3_rank5", FOUR -> ""
var StyleSuffix = map[string]string{
	"FOUR": "",
	"NS":   "_rank5",
	"AS":   "_s2_rank5",
	"ES":   "_s3_rank5",
}

// Stellar awakening image suffix, appended after style suffix.
const StellarSuffix = "_opened"

// Names for no spoiler. Key is the name when encounters at first, and the value is the true name for same character.
var SPOILER_NAMES = map[string]string{
	"Red Clad Flam.": "Dewey (Alter)",
	"Silver Striker": "Premaya (Alter)",
	"Violet Lancer":  "Toova (Alter)",
	"Cyan Scyther":   "Suzette (Alter)",
	"Black Clad Sw.": "Isuka (Alter)",
	"Azure Retainer": "Vares",
	"Nekoko":         "Necoco (Alter)",
	"Chelika":        "Spirika",
	"Dark Devourer":  "Mighty (Alter)",
}

const AEWIKI_ALTER_SUFFIX = "_(Alter)"
