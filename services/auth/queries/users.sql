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
