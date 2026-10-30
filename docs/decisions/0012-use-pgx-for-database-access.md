# 0012. Use pgx for database access, sqlc later

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

The Go services need to send SQL queries to Postgres. In milestone 1 there is only one query: read the greeting.

## Decision

Use pgx, the standard Postgres driver for Go, directly. Add sqlc (which generates Go code from SQL queries) in milestone 2, when there are many queries.

## Why

- One query only needs one short function. That is the simplest code that works.
- pgx is fast, well maintained and the most used Postgres driver for Go.
- sqlc builds on pgx, so nothing is thrown away when we add it.

## Alternatives considered

- **pgx + sqlc now:** the full pattern, but a whole code generation tool for one query.
- **database/sql + pgx:** Go's general database interface, an extra layer we don't need when we only use Postgres.
- **GORM or Ent (ORMs):** they hide the SQL, so you learn less and see less.

## Consequences

Milestone 2 adds sqlc and moves the queries into `.sql` files.
