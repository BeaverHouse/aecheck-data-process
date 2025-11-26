-- name: GetPersonalityMappings :many
SELECT
    p.personality_id,
    t.en as english_name
FROM aecheck.personality_mappings p
LEFT JOIN aecheck.translations t ON p.personality_id = t.key
WHERE p.character_id = $1
  AND p.deleted_at IS NULL;

-- name: SoftDeletePersonalityMappings :exec
UPDATE aecheck.personality_mappings
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE character_id = $1
  AND deleted_at IS NULL;

-- name: InsertPersonalityMapping :exec
INSERT INTO aecheck.personality_mappings (character_id, personality_id, description, created_at)
VALUES ($1, $2, NULL, CURRENT_TIMESTAMP);

-- name: PurgeDeletedPersonalityMappings :exec
DELETE FROM aecheck.personality_mappings
WHERE character_id = $1
  AND deleted_at IS NOT NULL;
