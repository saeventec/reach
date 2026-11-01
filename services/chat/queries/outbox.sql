-- name: CreateOutboxEvent :exec
-- Saves an event to publish. Runs inside the transaction of the change.
INSERT INTO outbox (subject, payload)
VALUES (sqlc.arg(subject), sqlc.arg(payload));

-- name: ListUnpublishedEvents :many
-- The events still waiting, the oldest first. Uses the partial index.
SELECT id, subject, payload
FROM outbox
WHERE published_at IS NULL
ORDER BY id ASC
LIMIT sqlc.arg(maximum_count);

-- name: MarkEventPublished :exec
UPDATE outbox
SET published_at = now()
WHERE id = sqlc.arg(id);
