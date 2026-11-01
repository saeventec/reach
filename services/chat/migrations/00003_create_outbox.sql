-- The outbox: events waiting to be published on NATS (ADR 0031). An event
-- is written here in the same transaction as the change it announces, so
-- either both are saved, or neither. The relay (outbox/relay.go) publishes
-- the rows and marks them published.
-- Learn more: docs/learn/transactional-outbox.md

-- +goose Up
CREATE TABLE outbox (
    -- 1, 2, 3 ... made by the database ("identity"). The relay publishes
    -- the events in this order.
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    -- The NATS subject, like "chat.message.sent".
    subject text NOT NULL,

    -- The event: a Protobuf message from proto/chat/v1/events.proto, in
    -- its binary form. bytea is Postgres's type for raw bytes.
    payload bytea NOT NULL,

    created_at timestamptz NOT NULL DEFAULT now(),

    -- Empty (NULL) until the relay has published the event.
    published_at timestamptz
);

-- A "partial index": it covers only the rows still waiting. The relay asks
-- for them every second, and this keeps the question fast even when
-- millions of published rows pile up.
CREATE INDEX outbox_unpublished ON outbox (id) WHERE published_at IS NULL;

-- +goose Down
DROP TABLE outbox;
