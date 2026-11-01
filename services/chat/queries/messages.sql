-- name: IsMember :one
-- Says whether the user is a member of the conversation. EXISTS stops at
-- the first matching row, and answers true or false.
SELECT EXISTS (
    SELECT 1
    FROM members
    WHERE conversation_id = sqlc.arg(conversation_id)
      AND user_id = sqlc.arg(user_id)
);

-- name: TakeNextSequence :one
-- Adds 1 to the conversation's counter and returns the new number. The
-- UPDATE locks the conversation's row until the transaction ends, so two
-- messages sent at the same moment wait for each other, and get 42 and
-- 43, never two 42s. It also moves the conversation to the top of the list.
UPDATE conversations
SET last_sequence = last_sequence + 1,
    last_activity_at = now()
WHERE id = sqlc.arg(conversation_id)
RETURNING last_sequence;

-- name: FindMessage :one
-- Finds a message by its id: to answer a retried send.
SELECT id, conversation_id, sequence, sender_id, text, sent_at
FROM messages
WHERE id = sqlc.arg(id);

-- name: CreateMessage :one
INSERT INTO messages (id, conversation_id, sequence, sender_id, text)
VALUES (sqlc.arg(id), sqlc.arg(conversation_id), sqlc.arg(sequence), sqlc.arg(sender_id), sqlc.arg(text))
RETURNING id, conversation_id, sequence, sender_id, text, sent_at;

-- name: ListMessagesAfter :many
-- The messages after a sequence number, the oldest first: to catch up.
SELECT id, conversation_id, sequence, sender_id, text, sent_at
FROM messages
WHERE conversation_id = sqlc.arg(conversation_id)
  AND sequence > sqlc.arg(after_sequence)
ORDER BY sequence ASC
LIMIT sqlc.arg(maximum_count);

-- name: ListMessagesBefore :many
-- The messages just before a sequence number, the newest first (the Go
-- code turns the order around): to scroll back.
SELECT id, conversation_id, sequence, sender_id, text, sent_at
FROM messages
WHERE conversation_id = sqlc.arg(conversation_id)
  AND sequence < sqlc.arg(before_sequence)
ORDER BY sequence DESC
LIMIT sqlc.arg(maximum_count);

-- name: ListNewestMessages :many
-- The newest messages, the newest first (the Go code turns the order
-- around): to open a conversation.
SELECT id, conversation_id, sequence, sender_id, text, sent_at
FROM messages
WHERE conversation_id = sqlc.arg(conversation_id)
ORDER BY sequence DESC
LIMIT sqlc.arg(maximum_count);

-- name: MarkDelivered :one
-- Moves the member's delivered watermark up. LEAST(…) cuts a number past
-- the newest message down to the newest one, so a buggy app can't mark
-- future messages. The last WHERE line makes it change only if the number
-- really goes up: then it returns the watermarks; otherwise it returns no
-- row. The UPDATE locks the member's row, so two calls at the same moment
-- take turns, and the second one sees the first one's number.
UPDATE members
SET delivered_up_to = LEAST(sqlc.arg(sequence), conversations.last_sequence)
FROM conversations
WHERE members.conversation_id = conversations.id
  AND members.conversation_id = sqlc.arg(conversation_id)
  AND members.user_id = sqlc.arg(user_id)
  AND LEAST(sqlc.arg(sequence), conversations.last_sequence) > members.delivered_up_to
RETURNING members.delivered_up_to, members.read_up_to;

-- name: MarkRead :one
-- Like MarkDelivered, for the read watermark. Read also means received, so
-- it moves the delivered watermark up with it (GREATEST: never down).
UPDATE members
SET read_up_to = LEAST(sqlc.arg(sequence), conversations.last_sequence),
    delivered_up_to = GREATEST(members.delivered_up_to, LEAST(sqlc.arg(sequence), conversations.last_sequence))
FROM conversations
WHERE members.conversation_id = conversations.id
  AND members.conversation_id = sqlc.arg(conversation_id)
  AND members.user_id = sqlc.arg(user_id)
  AND LEAST(sqlc.arg(sequence), conversations.last_sequence) > members.read_up_to
RETURNING members.delivered_up_to, members.read_up_to;
