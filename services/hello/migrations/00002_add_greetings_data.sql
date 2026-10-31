-- +goose Up

INSERT INTO greetings (message) VALUES ('Hello, World!');

-- +goose Down
DELETE FROM greetings WHERE message = 'Hello, World!';
