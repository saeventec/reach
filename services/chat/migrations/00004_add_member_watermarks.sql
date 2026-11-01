-- Delivered and read, as two "watermarks" per member (ADR 0032): the
-- sequence number up to which the member's devices received the messages,
-- and up to which the member read them.
-- Learn more: docs/learn/message-ordering.md#delivered-and-read

-- +goose Up
ALTER TABLE members
    -- 0 means none yet. Existing members start at 0 too.
    ADD COLUMN delivered_up_to bigint NOT NULL DEFAULT 0,
    ADD COLUMN read_up_to bigint NOT NULL DEFAULT 0,
    -- Reading a message means it was received, so read is never ahead.
    ADD CONSTRAINT members_watermarks CHECK (0 <= read_up_to AND read_up_to <= delivered_up_to);

-- +goose Down
ALTER TABLE members
    DROP CONSTRAINT members_watermarks,
    DROP COLUMN read_up_to,
    DROP COLUMN delivered_up_to;
