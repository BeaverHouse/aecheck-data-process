-- name: GetCharacterWithTranslation :one
SELECT
    t.en as english_name,
    c.style,
    c.light_shadow,
    c.category,
    c.max_manifest,
    c.is_awaken,
    c.is_alter,
    c.alter_character,
    c.aewiki_url,
    c.seesaa_url,
    TO_CHAR(c.update_date::timestamp, 'YYYY-MM-DD') as update_date,
    c.custom_manifest
FROM aecheck.characters c
LEFT JOIN aecheck.translations t ON c.character_code = t.key
WHERE c.character_id = $1;

-- name: GetCharacterCodeByWikiURL :one
SELECT character_code
FROM aecheck.characters
WHERE aewiki_url = $1
LIMIT 1;

-- name: GetCharacterJSONBData :one
SELECT
    personalities_data,
    dungeons_data,
    buddy_data
FROM aecheck.characters
WHERE character_id = $1;

-- name: CheckCharacterExists :one
SELECT EXISTS(
    SELECT 1 FROM aecheck.characters WHERE character_id = $1
) as exists;

-- name: UpdateCharacter :exec
UPDATE aecheck.characters
SET
    character_code = $2,
    category = $3,
    style = $4,
    light_shadow = $5,
    max_manifest = $6,
    is_awaken = $7,
    is_alter = $8,
    alter_character = $9,
    aewiki_url = $10,
    seesaa_url = $11,
    update_date = $12,
    custom_manifest = $13,
    personalities_data = $14,
    dungeons_data = $15,
    buddy_data = $16,
    updated_at = CURRENT_TIMESTAMP
WHERE character_id = $1;

-- name: InsertCharacter :exec
INSERT INTO aecheck.characters (
    character_id,
    character_code,
    category,
    style,
    light_shadow,
    max_manifest,
    is_awaken,
    is_alter,
    alter_character,
    aewiki_url,
    seesaa_url,
    update_date,
    custom_manifest,
    personalities_data,
    dungeons_data,
    buddy_data,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- name: GetFirstCharacterByCode :one
SELECT
    character_id,
    style,
    updated_at
FROM aecheck.characters
WHERE character_code = $1
ORDER BY character_id ASC
LIMIT 1;

-- name: CopyCharacterToNewID :exec
INSERT INTO aecheck.characters (
    character_id,
    character_code,
    category,
    style,
    light_shadow,
    max_manifest,
    is_awaken,
    is_alter,
    alter_character,
    aewiki_url,
    seesaa_url,
    update_date,
    custom_manifest,
    personalities_data,
    dungeons_data,
    buddy_data,
    created_at,
    updated_at
)
SELECT
    $1,
    c.character_code,
    c.category,
    c.style,
    c.light_shadow,
    c.max_manifest,
    c.is_awaken,
    c.is_alter,
    c.alter_character,
    c.aewiki_url,
    c.seesaa_url,
    c.update_date,
    c.custom_manifest,
    c.personalities_data,
    c.dungeons_data,
    c.buddy_data,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM aecheck.characters c
WHERE c.character_id = $2;

-- name: UpdateCharacterToFourStar :exec
UPDATE aecheck.characters
SET
    style = '☆4',
    max_manifest = 0,
    is_awaken = false,
    is_alter = false,
    custom_manifest = false,
    updated_at = CURRENT_TIMESTAMP
WHERE character_id = $1;
