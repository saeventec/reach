# 0032. Order messages by a sequence number per conversation

- **Status:** Accepted (proposed by the agent, from the brief's design rules)
- **Date:** 2026-11-01

## Context

Everyone in a conversation must see messages in the same order, retries must not create copies, and a device that was offline must catch up on exactly what it missed. The brief already asks for client-made UUIDv7 ids and a sequence number per conversation.

## Decision

- Each conversation has a `last_sequence` counter. Sending a message runs `UPDATE … SET last_sequence = last_sequence + 1 … RETURNING` and inserts the message with that number, in one transaction.
- The browser makes each message's id (UUIDv7). Sending an id that exists returns the stored message (idempotency).
- Delivered and read are two numbers per member ("watermarks"): `delivered_up_to` and `read_up_to`. They only go up.

## Why

- The `UPDATE` locks the conversation's row, so two messages at the same moment get 42 and 43, never two 42s.
- "Everything after 41" is a simple, exact way to catch up.
- One watermark per member is much smaller than one status row per message per person.

## Alternatives considered

- **Order by time:** clocks on different computers disagree, and two messages can have the same time.
- **One global Postgres sequence:** no lock, but gaps and no order per conversation.
- **A status row per message and person:** precise, but many rows and many writes.

## Consequences

- A very busy group waits on its row lock. Not a problem for friends; something to measure with load tests later.
- Read status is per person, not per device. The data model can add per-device later.
