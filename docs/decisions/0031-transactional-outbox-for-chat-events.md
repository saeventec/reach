# 0031. Publish chat events with a transactional outbox

- **Status:** Accepted (the owner's choice)
- **Date:** 2026-11-01

## Context

When a message is saved, `chat` must also publish an event on NATS, so `realtime` can push it to browsers. Postgres and NATS are two systems: a crash between "saved" and "published" would leave a message nobody was told about.

## Decision

`chat` saves each event in an `outbox` table, in the same database transaction as the change itself. A small loop in `chat` (the relay) publishes unpublished rows to NATS JetStream and marks them published. It is woken right after each change, and also checks every second.

## Why

- Either the change and its event are both saved, or neither. Nothing is "saved but never announced".
- A classic distributed-systems pattern, listed in the brief's concepts to practice.

## Alternatives considered

- **Publish right after the commit:** simplest, but a crash in between loses the event.
- **Read Postgres's change log (CDC, for example Debezium):** no outbox code, but a big extra system.

## Consequences

- Events are delivered "at least once": the relay can publish one twice (published, then crashed before marking it). Every receiver must ignore duplicates; the browser does it by sequence number.
- The publishing code sits behind a small interface, so Kafka can replace NATS later.
- Published rows pile up. For now we keep them; deleting old rows is a small later job.
