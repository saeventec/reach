-- The sessions table: one row per login. The browser keeps a random token
-- in a cookie; this table keeps only the token's hash (ADR 0022).
-- Learn more: docs/learn/sessions.md

-- +goose Up
CREATE TABLE sessions (
    id uuid PRIMARY KEY DEFAULT uuidv7(),

    -- REFERENCES: a "foreign key". Every session must belong to an existing
    -- user. ON DELETE CASCADE: deleting a user also deletes their sessions.
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,

    -- The SHA-256 hash of the token, as raw bytes (bytea). UNIQUE also
    -- creates an index, so finding a session by its hash is fast.
    token_hash bytea NOT NULL UNIQUE,

    created_at timestamptz NOT NULL DEFAULT now(),

    -- When the session ends. Each use pushes it 30 days ahead.
    expires_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE sessions;
