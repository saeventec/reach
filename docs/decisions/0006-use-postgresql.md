# 0006. Use PostgreSQL

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The services need a reliable database for accounts, conversations and messages.

## Decision

Use PostgreSQL in one container. Each service owns its own database (or schema) inside it, and no service reads another service's tables. Database changes are versioned SQL files run by goose.

## Why

- Reliable, free and very popular, with great documentation.
- Plain SQL is a skill that lasts a whole career.
- One container keeps the local setup light, while separate databases keep services independent.

## Alternatives considered

- **MySQL / MariaDB:** also solid, slightly fewer features.
- **MongoDB:** document database, less fitting for related data like users and conversations.
- **CockroachDB:** distributed Postgres-like database, too much for one computer.

## Consequences

Services can't join data across each other's tables. They share data through events or API calls instead.
