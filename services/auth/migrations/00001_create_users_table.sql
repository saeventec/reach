-- The first auth migration: the users table, one row per account.
-- Learn more: docs/learn/database-migrations.md and
-- docs/learn/postgresql.md#ids-and-constraints

-- +goose Up
CREATE TABLE users (
    -- uuidv7() is built into Postgres 18. A UUIDv7 starts with the time it
    -- was created, so new ids sort after old ones.
    id uuid PRIMARY KEY DEFAULT uuidv7(),

    -- UNIQUE: no two accounts can have the same username.
    username text NOT NULL UNIQUE,

    -- The argon2id hash of the password (step 4). Never the password itself.
    password_hash text NOT NULL,

    -- timestamptz: a moment in time, stored in UTC, the same everywhere.
    created_at timestamptz NOT NULL DEFAULT now(),

    -- The username rules (ADR 0021). The database refuses any row that
    -- breaks them, even if the Go code has a bug:
    -- - 3 to 32 characters long
    -- - only a-z, 0-9 and the separators . _ -
    -- - starts and ends with a letter or digit, never two separators in a row
    CONSTRAINT username_length CHECK (char_length(username) BETWEEN 3 AND 32),
    CONSTRAINT username_characters CHECK (username ~ '^[a-z0-9]([._-]?[a-z0-9])*$')
);

-- +goose Down
DROP TABLE users;
