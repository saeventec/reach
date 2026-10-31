-- Each session is one device: the browser it was made in, and when it was
-- last used. Learn more: docs/learn/sessions.md#devices

-- +goose Up
-- The browser's User-Agent header, as sent at login. DEFAULT '' fills the
-- column for sessions that existed before this migration.
ALTER TABLE sessions ADD COLUMN user_agent text NOT NULL DEFAULT '';

-- When the session was last used. Updated on every use.
ALTER TABLE sessions ADD COLUMN last_used_at timestamptz NOT NULL DEFAULT now();

-- An index makes "all sessions of this user" fast, without reading the
-- whole table. Listing devices does exactly that.
CREATE INDEX sessions_user_id_index ON sessions (user_id);

-- +goose Down
DROP INDEX sessions_user_id_index;
ALTER TABLE sessions DROP COLUMN last_used_at;
ALTER TABLE sessions DROP COLUMN user_agent;
