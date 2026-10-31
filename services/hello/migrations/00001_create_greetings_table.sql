-- The first migration: create the greetings table and add one greeting.
-- goose runs the "Up" part to apply it, and the "Down" part to undo it.
-- Learn more: docs/learn/database-migrations.md

-- +goose Up
CREATE TABLE greetings (
    -- A number that Postgres picks for each new row: 1, 2, 3, ...
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    message text NOT NULL
);

INSERT INTO greetings (message) VALUES ('Hello from Postgres!');

-- +goose Down
DROP TABLE greetings;
