-- The first chat migration: conversations and their members.
-- Learn more: docs/design/m03-real-time-chat.md#2-where-the-data-lives-the-chat-service-and-its-database

-- +goose Up
CREATE TABLE conversations (
    id uuid PRIMARY KEY DEFAULT uuidv7(),

    -- 'direct' (1:1) or 'group'.
    kind text NOT NULL,

    -- Only for groups: the name everyone sees.
    group_name text,

    -- Only for 1:1 chats: the two user ids, sorted and joined with a ":".
    -- UNIQUE means there can never be two 1:1 chats for the same pair, even
    -- if both people press "start chat" at the same moment. (UNIQUE allows
    -- many NULLs, so groups are not affected.)
    direct_pair_key text UNIQUE,

    created_at timestamptz NOT NULL DEFAULT now(),

    -- When something last happened. The list of conversations is sorted by
    -- it. New messages (the next migration) will move it.
    last_activity_at timestamptz NOT NULL DEFAULT now(),

    -- The database's own safety net for the rules in the Go code: a 1:1
    -- chat has a pair key and no name, a group has a name (1 to 64
    -- characters) and no pair key.
    CONSTRAINT conversation_kind CHECK (kind IN ('direct', 'group')),
    CONSTRAINT direct_has_pair_key CHECK (
        (kind = 'direct' AND direct_pair_key IS NOT NULL AND group_name IS NULL)
        OR (kind = 'group' AND group_name IS NOT NULL AND direct_pair_key IS NULL)
    ),
    CONSTRAINT group_name_length CHECK (char_length(group_name) BETWEEN 1 AND 64)
);

CREATE TABLE members (
    -- ON DELETE CASCADE: deleting a conversation deletes its members.
    conversation_id uuid NOT NULL REFERENCES conversations (id) ON DELETE CASCADE,

    -- The user's id from the auth service. Not a foreign key: the users
    -- table is in another database, which chat can't see (each service
    -- owns its data).
    user_id uuid NOT NULL,

    joined_at timestamptz NOT NULL DEFAULT now(),

    -- A "composite" primary key, made of two columns: a person is in a
    -- conversation at most once.
    PRIMARY KEY (conversation_id, user_id)
);

-- "All conversations of this user" is the most common question. The
-- primary key above starts with conversation_id, so it can't answer that
-- fast; this index starts with user_id.
CREATE INDEX members_user_id_index ON members (user_id);

-- +goose Down
DROP TABLE members;
DROP TABLE conversations;
