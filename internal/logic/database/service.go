package database

import (
	"aecheck-data-process/internal/db/postgres"
	"aecheck-data-process/internal/logic"
	"aecheck-data-process/internal/types"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool    *pgxpool.Pool
	queries *postgres.Queries
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		pool:    pool,
		queries: postgres.New(pool),
	}
}

func (s *Service) GetTranslationsByJapaneseList(ctx context.Context, jas []string) ([]postgres.GetTranslationsByJapaneseListRow, error) {
	return s.queries.GetTranslationsByJapaneseList(ctx, jas)
}

func (s *Service) GetTranslation(ctx context.Context, key string) (postgres.GetTranslationRow, error) {
	return s.queries.GetTranslation(ctx, key)
}

func (s *Service) UpdateTranslationValues(ctx context.Context, key, ko, en, ja string) error {
	return s.queries.UpdateTranslation(ctx, postgres.UpdateTranslationParams{
		Key: key,
		Ko:  ko,
		En:  en,
		Ja:  ja,
	})
}

func (s *Service) UpdateCharacterTier(ctx context.Context, characterID string, tier pgtype.Text) error {
	return s.queries.UpdateCharacterTier(ctx, postgres.UpdateCharacterTierParams{
		CharacterID: characterID,
		Tier:        tier,
	})
}

func (s *Service) ClearAllTiers(ctx context.Context) error {
	return s.queries.ClearAllTiers(ctx)
}

func (s *Service) GetCharacterByJapaneseName(ctx context.Context, ja string, style string) (postgres.GetCharacterByJapaneseNameRow, error) {
	return s.queries.GetCharacterByJapaneseName(ctx, postgres.GetCharacterByJapaneseNameParams{
		Ja:    ja,
		Style: style,
	})
}

func (s *Service) GetCharacterByEnglishName(ctx context.Context, en string, style string, isAlter bool) (postgres.GetCharacterByEnglishNameRow, error) {
	return s.queries.GetCharacterByEnglishName(ctx, postgres.GetCharacterByEnglishNameParams{
		En:      en,
		Style:   style,
		IsAlter: isAlter,
	})
}

func (s *Service) GetCharacterNumericID(ctx context.Context, gameID int, style string) (int, error) {
	charID, err := s.queries.GetCharacterByCodeAndStyle(ctx, postgres.GetCharacterByCodeAndStyleParams{
		CharacterCode: fmt.Sprintf("c%d", gameID),
		Style:         style,
	})
	if err != nil {
		return 0, err
	}
	n, err := strconv.Atoi(strings.Replace(charID, "char", "", 1))
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (s *Service) GetRandomCharacterWikiURL(ctx context.Context) (string, error) {
	url, err := s.queries.GetRandomCharacterWikiURL(ctx)
	if err != nil {
		return "", err
	}
	return url.String, nil
}

func (s *Service) GetRandomBuddyWikiURL(ctx context.Context) (string, error) {
	url, err := s.queries.GetRandomBuddyWikiURL(ctx)
	if err != nil {
		return "", err
	}
	return url.String, nil
}

func (s *Service) FindTranslationFromDB(ctx context.Context, englishName string, isClass bool) (string, *types.TranslationInfo, error) {
	searchName := logic.ResolveForDB(englishName)
	if isClass {
		row, err := s.queries.GetClassTranslationByEnglishName(ctx, searchName)
		if err != nil {
			return "", nil, err
		}
		return row.Key, &types.TranslationInfo{EnglishName: row.En, KoreanName: row.Ko, JapaneseName: row.Ja}, nil
	}
	row, err := s.queries.GetCharacterTranslationByEnglishName(ctx, searchName)
	if err != nil {
		return "", nil, err
	}
	return row.Key, &types.TranslationInfo{EnglishName: row.En, KoreanName: row.Ko, JapaneseName: row.Ja}, nil
}
