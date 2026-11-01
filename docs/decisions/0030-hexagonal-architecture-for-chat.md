# 0030. Hexagonal architecture for the chat service

- **Status:** Accepted (the owner's choice)
- **Date:** 2026-11-01

## Context

`chat` has real business rules: only members may send or read, there is only one 1:1 conversation per pair, read status only moves forward. In a flat folder (like `auth`), these rules get mixed with SQL, Connect and NATS code.

## Decision

`chat` uses hexagonal architecture ("ports and adapters"), kept small. A core package (`conversation/`) holds the rules and the interfaces it needs ("ports"). Separate packages ("adapters") implement them: `postgres/`, `api/` (Connect), `outbox/` (NATS) and `authclient/`. `main.go` connects them. See the [design doc](../design/m03-real-time-chat.md#10-code-structure-of-chat).

## Why

- The core can be read and tested without a database or a network.
- A part can be swapped without touching the rules, for example NATS for Kafka (milestone 12).
- It is on the [learning map](../PROJECT_BRIEF.md#learning-map) (clean architecture), and `auth` stays flat, so the two styles can be compared.

## Alternatives considered

- **Flat, like `auth`:** less code, but the rules mix with SQL and transport code.
- **Layered (handler → service → repository):** common and simple, but the core still depends on the database code.

## Consequences

- More folders and interfaces than `auth`. Each interface must earn its place: only the ones the core really needs.
- `realtime` has almost no business rules, so it stays flat.
