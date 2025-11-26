-- name: GetBuddyWithDetails :one
SELECT
    t.en as english_name,
    c.style,
    c.aewiki_url as partner_aewiki_url,
    b.aewiki_url,
    b.seesaa_url
FROM aecheck.buddies b
LEFT JOIN aecheck.characters c ON b.character_id = c.character_id
LEFT JOIN aecheck.translations t ON c.character_code = t.key
WHERE b.buddy_id = $1;

-- name: GetCharacterIDByWikiURL :one
SELECT character_id
FROM aecheck.characters
WHERE aewiki_url = $1
LIMIT 1;

-- name: CheckBuddyExists :one
SELECT EXISTS(
    SELECT 1 FROM aecheck.buddies WHERE buddy_id = $1
) as exists;

-- name: UpdateBuddyWithCharacter :exec
UPDATE aecheck.buddies
SET
    character_id = $2,
    aewiki_url = $3,
    seesaa_url = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE buddy_id = $1;

-- name: UpdateBuddyWithGetPath :exec
UPDATE aecheck.buddies
SET
    get_path = $2,
    aewiki_url = $3,
    seesaa_url = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE buddy_id = $1;

-- name: InsertBuddyWithCharacter :exec
INSERT INTO aecheck.buddies (buddy_id, character_id, aewiki_url, seesaa_url, created_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);

-- name: InsertBuddyWithGetPath :exec
INSERT INTO aecheck.buddies (buddy_id, get_path, aewiki_url, seesaa_url, created_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP);
