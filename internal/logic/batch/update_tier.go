package batch

import (
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/logic/common"
	"aecheck-data-process/internal/logic/database"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/BeaverHouse/go-common/logger"
	"github.com/jackc/pgx/v5/pgtype"
)

// TierCharacter represents a character from tier list
type TierCharacter struct {
	Name    string `json:"name"`     // Japanese or English name
	Style   string `json:"style"`    // "NS", "AS", "ES"
	IsAlter bool   `json:"is_alter"` // true for AC characters
	Lang    string `json:"lang"`     // "ja" or "en"
}

// TierInput represents the JSON input from Claude Skill (site-based)
type TierInput struct {
	Altema      []TierCharacter `json:"altema"`
	Seesaa      []TierCharacter `json:"seesaa"`
	AnotherTier []TierCharacter `json:"anothertier"`
}

// resolvedCharacter holds resolved character info
type resolvedCharacter struct {
	CharacterID string
	KoreanName  string
	Style       string
	IsAlter     bool
}

// UpdateTierFromJSON reads tier data from JSON file and updates DB
func UpdateTierFromJSON(jsonPath string, dbService *database.Service, dryrun bool) {
	common.Log.Info("Update Tier from JSON")

	// Read JSON file
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		common.Log.Error("Error reading JSON file", logger.Field{Key: "error", Value: err})
		return
	}

	var input TierInput
	if err := json.Unmarshal(data, &input); err != nil {
		common.Log.Error("Error parsing JSON", logger.Field{Key: "error", Value: err})
		return
	}

	ctx := context.Background()

	// Resolve each site's characters
	common.Log.Info("[Altema 99점]")
	altemaResolved := resolveCharacters(ctx, input.Altema, dbService)

	common.Log.Info("[Seesaa EXC/SSS]")
	seesaaResolved := resolveCharacters(ctx, input.Seesaa, dbService)

	common.Log.Info("[AnotherTier Best/SSS/SS]")
	anotherResolved := resolveCharacters(ctx, input.AnotherTier, dbService)

	// Calculate overlaps
	common.Log.Info("Tier 계산")

	// Count occurrences by character_id
	counts := make(map[string]int)
	charInfo := make(map[string]resolvedCharacter)

	for _, c := range altemaResolved {
		counts[c.CharacterID]++
		charInfo[c.CharacterID] = c
	}
	for _, c := range seesaaResolved {
		counts[c.CharacterID]++
		charInfo[c.CharacterID] = c
	}
	for _, c := range anotherResolved {
		counts[c.CharacterID]++
		charInfo[c.CharacterID] = c
	}

	// Categorize
	var superOP, op []resolvedCharacter
	for charID, count := range counts {
		info := charInfo[charID]
		if count >= 3 {
			superOP = append(superOP, info)
		} else if count == 2 {
			op = append(op, info)
		}
	}

	// Display results
	common.Log.Info("Super OP (3개 사이트)", logger.Field{Key: "count", Value: len(superOP)})
	for _, c := range superOP {
		alter := ""
		if c.IsAlter {
			alter = " [AC]"
		}
		fmt.Printf("  - %s (%s)%s\n", c.KoreanName, c.Style, alter)
	}

	common.Log.Info("OP (2개 사이트)", logger.Field{Key: "count", Value: len(op)})
	for _, c := range op {
		alter := ""
		if c.IsAlter {
			alter = " [AC]"
		}
		fmt.Printf("  - %s (%s)%s\n", c.KoreanName, c.Style, alter)
	}

	if dryrun {
		common.Log.Info("[DRYRUN] DB 업데이트 생략")
		return
	}

	// Ask user confirmation
	fmt.Print("이 tier 정보를 DB에 반영할까요? (y/N): ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "yes" {
		common.Log.Info("취소되었습니다")
		return
	}

	// Clear existing tiers
	common.Log.Info("기존 tier 초기화 중...")
	if err := dbService.ClearAllTiers(ctx); err != nil {
		common.Log.Error("Error clearing tiers", logger.Field{Key: "error", Value: err})
		return
	}

	// Update Super OP
	common.Log.Info("Super OP 업데이트 중...")
	for _, c := range superOP {
		if err := dbService.UpdateCharacterTier(ctx, c.CharacterID, pgtype.Text{String: "super_op", Valid: true}); err != nil {
			common.Log.Error("Error updating tier", logger.Field{Key: "name", Value: c.KoreanName}, logger.Field{Key: "error", Value: err})
		} else {
			common.Log.Info("Updated tier", logger.Field{Key: "name", Value: c.KoreanName}, logger.Field{Key: "tier", Value: "super_op"})
		}
	}

	// Update OP
	common.Log.Info("OP 업데이트 중...")
	for _, c := range op {
		if err := dbService.UpdateCharacterTier(ctx, c.CharacterID, pgtype.Text{String: "op", Valid: true}); err != nil {
			common.Log.Error("Error updating tier", logger.Field{Key: "name", Value: c.KoreanName}, logger.Field{Key: "error", Value: err})
		} else {
			common.Log.Info("Updated tier", logger.Field{Key: "name", Value: c.KoreanName}, logger.Field{Key: "tier", Value: "op"})
		}
	}

	common.Log.Info("완료!")
}

// resolveCharacters resolves character names to IDs and Korean names
func resolveCharacters(ctx context.Context, chars []TierCharacter, dbService *database.Service) []resolvedCharacter {
	var resolved []resolvedCharacter

	for _, c := range chars {
		var charID, koreanName string

		if c.Lang == "ja" {
			// Japanese name lookup
			char, e := dbService.GetCharacterByJapaneseName(ctx, c.Name, c.Style)
			if e != nil {
				common.Log.Warn("Character not found", logger.Field{Key: "name", Value: c.Name}, logger.Field{Key: "style", Value: c.Style})
				continue
			}
			charID = char.CharacterID

			// Get Korean name
			trans, e := dbService.GetTranslation(ctx, char.CharacterCode)
			if e != nil {
				koreanName = c.Name // fallback to Japanese
			} else {
				koreanName = trans.Ko
			}
		} else {
			// English name lookup
			searchName := logic.ResolveForDB(c.Name)

			char, e := dbService.GetCharacterByEnglishName(ctx, searchName, c.Style, c.IsAlter)
			if e != nil {
				common.Log.Warn("Character not found", logger.Field{Key: "name", Value: c.Name}, logger.Field{Key: "style", Value: c.Style})
				continue
			}
			charID = char.CharacterID

			// Get Korean name
			trans, e := dbService.GetTranslation(ctx, char.CharacterCode)
			if e != nil {
				koreanName = c.Name // fallback to English
			} else {
				koreanName = trans.Ko
			}
		}

		alter := ""
		if c.IsAlter {
			alter = " [AC]"
		}
		fmt.Printf("  %s (%s)%s\n", koreanName, c.Style, alter)

		resolved = append(resolved, resolvedCharacter{
			CharacterID: charID,
			KoreanName:  koreanName,
			Style:       c.Style,
			IsAlter:     c.IsAlter,
		})
	}

	return resolved
}
