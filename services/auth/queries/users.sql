-- name: CreateUser :one
-- Adds a new account. It returns the new row without the password hash,
-- so the hash never leaves this query by accident.
INSERT INTO users (username, password_hash)
VALUES ($1, $2)
RETURNING id, username, created_at;

-- name: FindUserByUsername :one
-- Finds an account for login. It is the only query that returns the
-- password hash.
SELECT id, username, password_hash
FROM users
WHERE username = $1;

-- name: FindUsersByUsernames :many
-- Finds the accounts with any of these usernames, for other services
-- (user_directory.go). "= ANY(array)" means "equals one of the array's
-- values". "::text[]" tells sqlc and Postgres that the argument is an
-- array of text. Only the public columns, never the password hash.
SELECT id, username
FROM users
WHERE username = ANY(sqlc.arg(usernames)::text[]);

-- name: FindUsersByIDs :many
-- Finds the accounts with any of these ids, for other services.
SELECT id, username
FROM users
WHERE id = ANY(sqlc.arg(user_ids)::uuid[]);
