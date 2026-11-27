-- name: TruncateSeedTables :exec
TRUNCATE aecheck.characters, aecheck.buddies, aecheck.dungeons, aecheck.translations;

-- name: RegisterServiceToken :exec
-- A fresh test database has no token for the AE Check web app, and the
-- analyzer refuses its requests without one.
INSERT INTO tinyclover.service_tokens (name, token, scopes)
VALUES ('e2e-aecheck', $1, '{aecheck,file}')
ON CONFLICT (token) DO UPDATE SET scopes = EXCLUDED.scopes;
