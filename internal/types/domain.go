package types

// AEStyle represents the style/rarity of a character
type AEStyle string

const (
	StyleNS   AEStyle = "NS"
	StyleAS   AEStyle = "AS"
	StyleES   AEStyle = "ES"
	StyleFOUR AEStyle = "FOUR"
)

// AELightShadow represents the light/shadow attribute
type AELightShadow string

const (
	LSLight  AELightShadow = "light"
	LSShadow AELightShadow = "shadow"
)

// AECategory represents the character category
type AECategory string

const (
	AECategoryEncounter AECategory = "ENCOUNTER" // 몽견
	AECategoryFree      AECategory = "FREE"      // 배포
	AECategoryColab     AECategory = "COLAB"     // 콜라보 (협주)
)

// CharacterInfoFromAEWikiURL contains basic character info from wiki URL
type CharacterInfoFromAEWikiURL struct {
	EnglishName string  // 영문 이름
	Style       AEStyle // 스타일
	IsAlter     bool    // 이시층 여부
}

// CharacterInfoFromAEWiki contains complete character info from wiki
type CharacterInfoFromAEWiki struct {
	CharacterInfoFromAEWikiURL
	GameID           int           // 게임 내 숫자로 된 ID
	EnglishClassName string        // 영문 클래스 이름
	IsAwaken         bool          // 성도각성 여부
	LightShadow      AELightShadow // 천/명 속성
	Category         AECategory    // 캐릭터 분류
	Personalities    []string      // 퍼스널리티
	UpdateDate       string        // 업데이트 날짜
	MaxManifest      int           // 최대 현현 단계
	IsManifestCustom bool          // 커스텀 현현 (Weapon Tempering) 여부
	Dungeons         []string      // 직업서 드랍 던전 (이절, 개전, 경전록은 던전이 없으므로 그냥 적음)
	WikiURL          string        // 영문 위키 링크
}

// BuddyInfoFromAEWiki contains buddy/companion info from wiki
type BuddyInfoFromAEWiki struct {
	GameID      int     // 게임 내 숫자로 된 ID
	EnglishName string  // 영문 이름
	Style       AEStyle // 스타일
	PartnerLink string  // 파트너 캐릭터 링크
	WikiURL     string  // 영문 위키 링크
}

// TranslationInfo contains translation data for multiple languages
type TranslationInfo struct {
	EnglishName  string // 영문 이름
	KoreanName   string // 한글 이름
	JapaneseName string // 일본어 이름
}

// UpdateStatus represents the status of a character update check
type UpdateStatus string

const (
	Updated    UpdateStatus = "updated"
	NotUpdated UpdateStatus = "not_updated"
	NotExists  UpdateStatus = "not_exists"
	NotNeeded  UpdateStatus = "not_needed"
)
