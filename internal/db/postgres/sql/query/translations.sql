-- name: GetTranslation :one
SELECT ko, en, ja
FROM aecheck.translations
WHERE key = $1;

-- name: CheckTranslationExists :one
SELECT EXISTS(
    SELECT 1 FROM aecheck.translations WHERE key = $1
) as exists;

-- name: UpdateTranslation :exec
UPDATE aecheck.translations
SET
    ko = $2,
    en = $3,
    ja = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE key = $1;

-- name: InsertTranslation :exec
INSERT INTO aecheck.translations (key, ko, en, ja, created_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);

-- name: GetKeyByEnglishName :one
SELECT key
FROM aecheck.translations
WHERE en = $1
LIMIT 1;

-- name: GetTranslationsByJapaneseList :many
SELECT key, ko, en, ja
FROM aecheck.translations
WHERE ja = ANY($1::text[]);

-- name: GetCharacterTranslationByEnglishName :one
SELECT key, ko, en, ja
FROM aecheck.translations
WHERE en = $1 AND key LIKE 'c%' AND key NOT LIKE 'char%'
LIMIT 1;

-- name: GetCharacterTranslationByExactEnglishName :one
SELECT key, ko, en, ja
FROM aecheck.translations
WHERE en = $1
    AND (key LIKE 'c%' OR key LIKE 'spoiler.c%')
    AND key NOT LIKE 'char%'
LIMIT 1;

-- name: GetClassTranslationByEnglishName :one
SELECT key, ko, en, ja
FROM aecheck.translations
WHERE en = $1 AND key LIKE 'book.char%'
LIMIT 1;
