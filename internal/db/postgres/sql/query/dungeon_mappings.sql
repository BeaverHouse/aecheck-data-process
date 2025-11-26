-- name: GetDungeonMappings :many
SELECT
    d.dungeon_id,
    t.en as english_name
FROM aecheck.dungeon_mappings d
LEFT JOIN aecheck.translations t ON d.dungeon_id = t.key
WHERE d.character_id = $1
  AND d.deleted_at IS NULL;

-- name: SoftDeleteDungeonMappings :exec
UPDATE aecheck.dungeon_mappings
SET
    deleted_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE character_id = $1
  AND deleted_at IS NULL;

-- name: InsertDungeonMapping :exec
INSERT INTO aecheck.dungeon_mappings (character_id, dungeon_id, description, created_at)
VALUES ($1, $2, NULL, CURRENT_TIMESTAMP);

-- name: PurgeDeletedDungeonMappings :exec
DELETE FROM aecheck.dungeon_mappings
WHERE character_id = $1
  AND deleted_at IS NOT NULL;
