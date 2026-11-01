-- Messages, and the counter that numbers them per conversation (ADR 0032).
-- Learn more: docs/design/m03-real-time-chat.md#3-order-a-sequence-number-per-conversation

-- +goose Up

-- The sequence number of the conversation's newest message. 0 means no
-- message yet. Sending a message adds 1 here and uses the new number.
ALTER TABLE conversations ADD COLUMN last_sequence bigint NOT NULL DEFAULT 0;

CREATE TABLE messages (
    -- Made by the app, not by the database (no DEFAULT): a retry sends the
    -- same id, so the PRIMARY KEY stops a copy from being stored.
    id uuid PRIMARY KEY,

    conversation_id uuid NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,

    -- The message's place in its conversation: 1, 2, 3 ... bigint is a
    -- 64-bit whole number, so it never runs out.
    sequence bigint NOT NULL,

    -- The sender's user id from the auth service.
    sender_id uuid NOT NULL,

    text text NOT NULL,

    sent_at timestamptz NOT NULL DEFAULT now(),

    -- Two messages can never share a place in one conversation. This also
    -- creates an index, which makes "messages after 41" and "messages
    -- before 80" fast.
    CONSTRAINT messages_conversation_sequence UNIQUE (conversation_id, sequence),
    CONSTRAINT messages_sequence_positive CHECK (sequence > 0),
    CONSTRAINT messages_text_length CHECK (char_length(text) BETWEEN 1 AND 4000)
);

-- +goose Down
DROP TABLE messages;
ALTER TABLE conversations DROP COLUMN last_sequence;
