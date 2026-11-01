-- name: CreateDirectConversation :one
-- Creates the 1:1 conversation for a pair, unless it exists already.
-- "ON CONFLICT ... DO NOTHING" means: if the UNIQUE pair key is taken, add
-- nothing and return no row (the caller then finds the existing one).
-- Asking first and inserting after could fail when two requests race.
INSERT INTO conversations (kind, direct_pair_key)
VALUES ('direct', sqlc.arg(direct_pair_key))
ON CONFLICT (direct_pair_key) DO NOTHING
RETURNING id;

-- name: FindDirectConversation :one
-- Finds the 1:1 conversation of a pair.
SELECT id
FROM conversations
WHERE direct_pair_key = sqlc.arg(direct_pair_key);

-- name: CreateGroupConversation :one
-- Creates a group.
INSERT INTO conversations (kind, group_name)
VALUES ('group', sqlc.arg(group_name))
RETURNING id;

-- name: AddMember :exec
-- Adds a person to a conversation. A person who is already in it stays
-- once (the primary key), so this is safe to repeat.
INSERT INTO members (conversation_id, user_id)
VALUES (sqlc.arg(conversation_id), sqlc.arg(user_id))
ON CONFLICT (conversation_id, user_id) DO NOTHING;

-- name: FindConversation :one
-- Finds one conversation by its id.
SELECT id, kind, group_name, created_at, last_activity_at
FROM conversations
WHERE id = sqlc.arg(id);

-- name: ListConversationsOfUser :many
-- The user's conversations, the most recently active first. The JOIN keeps
-- only conversations where this user is a member.
SELECT conversations.id, conversations.kind, conversations.group_name,
       conversations.created_at, conversations.last_activity_at
FROM conversations
JOIN members ON members.conversation_id = conversations.id
WHERE members.user_id = sqlc.arg(user_id)
ORDER BY conversations.last_activity_at DESC, conversations.id DESC;

-- name: ListMembersOfConversations :many
-- The members of several conversations at once, so listing 20
-- conversations takes one query, not 20.
SELECT conversation_id, user_id, delivered_up_to, read_up_to
FROM members
WHERE conversation_id = ANY(sqlc.arg(conversation_ids)::uuid[])
ORDER BY conversation_id, joined_at, user_id;

-- name: ListMemberIDs :many
-- The user ids of one conversation's members: for the events, which say
-- who to deliver them to.
SELECT user_id
FROM members
WHERE conversation_id = sqlc.arg(conversation_id)
ORDER BY joined_at, user_id;
