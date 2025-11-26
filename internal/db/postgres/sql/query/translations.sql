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
