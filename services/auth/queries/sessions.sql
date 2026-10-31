-- name: CreateSession :exec
-- Saves a new login. sqlc.arg(name) gives the Go parameter a clear name.
INSERT INTO sessions (user_id, token_hash, user_agent, expires_at)
VALUES (sqlc.arg(user_id), sqlc.arg(token_hash), sqlc.arg(user_agent), sqlc.arg(expires_at));

-- name: RenewSession :one
-- Finds the session with this token hash, if it has not expired yet, moves
-- its end to the new expiry and records that it was used, all in one step.
-- It returns the session and its user. No row means: no valid session.
UPDATE sessions
SET expires_at = sqlc.arg(new_expires_at),
    last_used_at = now()
FROM users
WHERE sessions.token_hash = sqlc.arg(token_hash)
  AND sessions.expires_at > now()
  AND users.id = sessions.user_id
RETURNING sessions.id AS session_id, users.id AS user_id, users.username;

-- name: ListSessions :many
-- All valid sessions of one user, the most recently used first.
SELECT id, user_agent, created_at, last_used_at
FROM sessions
WHERE user_id = sqlc.arg(user_id)
  AND expires_at > now()
ORDER BY last_used_at DESC;

-- name: DeleteSession :execrows
-- Ends one session. The user_id check means a user can only end their own
-- sessions, even if they send someone else's session id. ":execrows"
-- returns how many rows were deleted: 0 means there was no such session.
DELETE FROM sessions
WHERE id = sqlc.arg(session_id)
  AND user_id = sqlc.arg(user_id);
