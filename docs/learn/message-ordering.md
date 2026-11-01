# Message ordering and idempotency

## What it is

Two promises every chat app must keep:

- **Everyone sees the same order.** If alice sees "Hi" before "How are you?", so does bob, on every device.
- **A message is stored once,** even if the app sends it twice. The app sends again when the network drops before the answer arrives, because it can't know whether the first try arrived.

Reach keeps them with two simple tools: a **sequence number** per conversation (1, 2, 3 ...), and **ids made by the app** ([ADR 0032](../decisions/0032-order-messages-by-sequence-per-conversation.md)).

## Why we use it

- **Why not order by time:** two computers' clocks never agree exactly, and two messages can have the same time. The server's clock alone would work for one server, but not with several copies of `chat` (milestone 12).
- **Catching up becomes one question:** a device that was offline asks "everything after 41". No dates, no guessing.
- **The server never reads the text to do this.** Routing and order use only numbers and ids, so the text can be end-to-end encrypted later (milestone 8).

## The sequence number

Sending a message runs in one transaction ([services/chat/postgres/message_store.go](../../services/chat/postgres/message_store.go)):

```sql
UPDATE conversations
SET last_sequence = last_sequence + 1, last_activity_at = now()
WHERE id = $1
RETURNING last_sequence;             -- say, 42

INSERT INTO messages (id, conversation_id, sequence, ...) VALUES (..., 42, ...);
COMMIT;
```

- The `UPDATE` **locks** the conversation's row until `COMMIT`. A second message in the same conversation waits at its own `UPDATE`, and then gets 43. Never two 42s.
- The lock is per conversation: messages in other conversations don't wait.
- `UNIQUE (conversation_id, sequence)` is the database's safety net: even a bug can't store two 42s.
- **The cost:** a very busy group waits on one row. For friends, a wait of a few milliseconds at most.

## Idempotency: ids made by the app

**Idempotent** means: doing it twice has the same effect as doing it once. Pressing a lift button twice still calls one lift.

1. The app makes the message's id (a UUIDv7) **before** sending.
2. It sends. The network drops; the app doesn't know if it arrived.
3. It sends again, **with the same id**.
4. The server finds the id, and returns the stored message instead of storing a copy.

The server also checks that a found id belongs to the same sender, conversation and text. Otherwise someone is reusing an id: "this message id is already used".

**UUIDv7** starts with the time it was made, so newer ids sort after older ones, and an index on them stays in order. The server checks the version: the 13th digit of a UUIDv7 is `7`.

## When two retries collide

Usually the second try comes after the first is stored, and a simple lookup finds it. But two tries can arrive at the same moment:

```
try A: looks up the id → not found → waits for the lock
try B: looks up the id → not found → waits for the lock
try A: gets the lock → 42 → stores the message → COMMIT
try B: gets the lock → 43 → INSERT fails: the id exists (primary key)
try B: ROLLBACK (also undoes its +1, so 43 is free again) → looks up → returns A's message
```

So `id uuid PRIMARY KEY` is what really guarantees "stored once". The lookup before is only a shortcut.

## Testing a race for real

The first test for this sent 10 retries at once, and passed. But it also passed when the code for "try B" was removed: the tries never really collided, because the first one finished before the others even looked.

The fixed test makes the collision certain:

1. The test locks the conversation's row itself: `SELECT ... FOR UPDATE` in its own transaction.
2. It starts 5 retries. All of them miss the lookup, and wait for the lock.
3. It waits until Postgres shows 5 waiting connections in `pg_stat_activity` (a built-in view: one row per connection, and what it is doing).
4. It commits, and all 5 continue at once.

Now the test fails with "duplicate key value violates unique constraint" when the code for "try B" is removed. **A test only counts once you have seen it fail for the bug it is meant to catch.**

## Alternatives

- **Timestamps:** simple, but clocks disagree and times can repeat.
- **One global Postgres sequence (`bigserial`):** no lock, but numbers skip (a rolled-back insert still uses its number), and there is no "per conversation" order.
- **Ids made by the server:** simpler, but a retry stores a copy.
- **An "idempotency key" header** (as payment APIs like Stripe use): the same idea, for any request; our message id plays that role.

## Learn more

- https://www.postgresql.org/docs/current/explicit-locking.html#LOCKING-ROWS
- https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW
- https://www.rfc-editor.org/rfc/rfc9562#name-uuid-version-7 (the UUIDv7 standard)
