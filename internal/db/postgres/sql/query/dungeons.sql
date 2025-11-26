-- name: GetDungeonByID :one
SELECT
    dungeon_id,
    altema_url,
    aewiki_url
FROM aecheck.dungeons
WHERE dungeon_id = $1;

-- name: GetDungeonsByIDs :many
SELECT
    dungeon_id,
    altema_url,
    aewiki_url
FROM aecheck.dungeons
WHERE dungeon_id = ANY($1::text[]);
