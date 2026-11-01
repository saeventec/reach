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

The browser's own `crypto.randomUUID()` makes version 4 (all random), so the web app makes its v7 ids itself, in [apps/web/src/chat/uuidv7.ts](../../apps/web/src/chat/uuidv7.ts): 6 bytes of time in milliseconds, 10 random bytes from `crypto.getRandomValues` (the browser's secure random generator), and two small markers for the version and variant. About 25 lines, so no library is needed. (Alternatives: the `uuid` npm package has `v7()`; newer browsers may add it one day.)

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

## Catching up in the app

The web app gets messages four ways: the first page, older pages, the answer to its own send, and live from the WebSocket. Live delivery can lose events (the WebSocket was closed) or repeat them ("at least once"). The sequence numbers make it safe anyway. [apps/web/src/chat/conversationSync.ts](../../apps/web/src/chat/conversationSync.ts) has the rules, in a plain TypeScript class (not React), so its tests need no page:

- **Every message goes through one merge** ([messageTimeline.ts](../../apps/web/src/chat/messageTimeline.ts)): kept by sequence number, sorted. A copy has the same number, so it simply replaces itself. Your own message arrives twice (the send's answer and the live event) and shows once.
- **A gap means something was missed.** With 1 to 42 on screen, a live message 45 is added, and then the app asks `ListMessages(after: 42)`, which brings 43 and 44 (and 45 again, ignored).
- **After every (re)connect** it asks for everything after the newest message, page by page, until `has_more` is false.
- **Only one catch-up at a time.** One asked for while another runs sets a flag, and the running one goes once more at the end. Otherwise two could ask "after 42" at once, which is only wasteful, or one could miss a message that arrived between its question and its answer, which is wrong. A test holds the first answer back to check this.
- **Before the first page arrives,** live messages are merged but don't count as gaps: there is nothing to compare them with yet.

The WebSocket only makes things fast. Correctness comes from `ListMessages(after: N)`.

## Delivered and read

Sequence numbers also make "delivered" (✓✓) and "read" small. Each member has two numbers, called **watermarks**, in the `members` table: `delivered_up_to` and `read_up_to`. "bob: read up to 42" means bob has read messages 1 to 42. Like the line a flood leaves on a wall: it says how high the water came, and everything below it is covered.

- One pair of numbers per member, not one row per message per person. A group of 50 with 10 000 messages needs 50 rows, not 500 000.
- `MarkDelivered(conversation, 42)`: the app received the messages up to 42 (live, or by catching up). `MarkRead(conversation, 42)`: they were on screen. Read also moves delivered, because you can't read what you didn't receive. A `CHECK` in the database makes sure `read_up_to` is never above `delivered_up_to`.
- **They only go up.** The query ([services/chat/queries/messages.sql](../../services/chat/queries/messages.sql)) changes the row only if the new number is higher (`… > members.delivered_up_to` in its `WHERE`). A late or repeated call changes nothing, so the calls are idempotent too.
- **Never past the newest message.** `LEAST(42, conversations.last_sequence)` cuts a too-high number down. Otherwise a buggy app could mark message 1000 "read" before it was even sent.
- **Two calls at the same moment:** the `UPDATE` locks the member's row, so they take turns. With Postgres's default isolation ("read committed"), the second one checks its `WHERE` again against the first one's new number. A test sends 30 calls at once with 1, 2 and 3: the result is always 3. Without the "only up" line, it ended at 2, with 30 events.
- **An event only when something changed:** the query returns a row only when the number moved. Then the store saves a `chat.receipt.updated` event, in the same transaction ([the outbox](transactional-outbox.md)). The event carries both numbers as they are now, so a receiver that missed one event is right again after the next.
- **When the app marks** ([useMarkAsSeen.ts](../../apps/web/src/chat/useMarkAsSeen.ts), [markDelivered.ts](../../apps/web/src/chat/markDelivered.ts)): an open conversation marks read up to its newest message while the tab is visible, and only delivered while the tab is in the background (read when you come back). The conversation list, and the page for messages in other chats, mark new messages from others delivered as they arrive, and when the list loads.
- **How the app shows it:** alice's message 42 is delivered to bob when bob's `delivered_up_to` is 42 or more, and read when his `read_up_to` is. In a group, ✓✓ means "everyone else's number is that high". `ListConversations` returns every member's watermarks, so the app knows them when it opens.

## Alternatives

- **Timestamps:** simple, but clocks disagree and times can repeat.
- **One global Postgres sequence (`bigserial`):** no lock, but numbers skip (a rolled-back insert still uses its number), and there is no "per conversation" order.
- **Ids made by the server:** simpler, but a retry stores a copy.
- **An "idempotency key" header** (as payment APIs like Stripe use): the same idea, for any request; our message id plays that role.

## Learn more

- https://www.postgresql.org/docs/current/explicit-locking.html#LOCKING-ROWS
- https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW
- https://www.rfc-editor.org/rfc/rfc9562#name-uuid-version-7 (the UUIDv7 standard)
