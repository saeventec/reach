-- name: CreateSession :exec
-- Saves a new login. sqlc.arg(name) gives the Go parameter a clear name.
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(expires_at));

-- name: RenewSession :one
-- Finds the session with this token hash, if it has not expired yet, and
-- moves its end to the new expiry, all in one step. It returns the session's
-- user. No row means: no valid session.
UPDATE sessions
SET expires_at = sqlc.arg(new_expires_at)
FROM users
WHERE sessions.token_hash = sqlc.arg(token_hash)
  AND sessions.expires_at > now()
  AND users.id = sessions.user_id
RETURNING users.id, users.username;
