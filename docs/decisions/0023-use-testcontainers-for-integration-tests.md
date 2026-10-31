# 0023. Use Testcontainers for integration tests

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Until now, tests use a fake store instead of Postgres. The `auth` service has many real SQL queries (sign up, find a user, sessions). A fake can't tell us whether the SQL itself is right, or whether a unique username is really enforced by the database.

## Decision

Add **integration tests** that run our queries against a real Postgres. The tests start Postgres in a Docker container themselves, with the Testcontainers library for Go, and remove it when they finish.

## Why

- The tests check the real SQL, the real migrations and the real database rules.
- Each test run starts with a clean, empty database, so tests don't depend on each other or on your local data.
- It only needs Docker, which we already have, on your computer and on GitHub's CI machines.

## Alternatives considered

- **Only fakes:** fast and simple, but they never test the SQL.
- **The Postgres from `task dev`:** no new library, but the tests would need it running, and could change or delete your local data.
- **A separate test Postgres in Compose:** works, but someone has to start it and keep it clean.

## Consequences

- Integration tests are slower (a few seconds to start Postgres) and need Docker running.
- Fast unit tests with fakes stay, for the logic that doesn't need a database.
- One new test-only library. We explain it when we add it (step 8).

**Update (2026-10-31), built in milestone 2, step 8:** `testcontainers-go` v0.44.0 with its Postgres module, and goose v3.29.0 as a library to apply the migrations (the same version as `auth-migrate`). One container per test run, with subtests. It brought about 130 indirect modules into `go.mod`, test-only. The auth image's `go mod download` step now also downloads them, which makes the first image build slower.
