# Milestone 2: Accounts

- **Status:** In progress (started 2026-10-31, on the branch `milestone-2-accounts`)

## Goal

People can sign up with a username and a password, log in, stay logged in, see the devices they are logged in on, and log out (also another device). The new `auth` service does this, with its own database.

```
Browser (React) → Caddy → auth service → Postgres (auth database)
```

## Steps

Each step is one small commit that runs, reviewed by the owner before the next one starts.

- [x] 1. Plan: this file and the decisions below
- [ ] 2. The `auth` service skeleton, built from the useful parts of `hello` (config, health check, Dockerfile, graceful shutdown). It gets its own database and its own database user. Compose and Caddy know about it, and `task lint` and `task test` cover both services.
- [ ] 3. sqlc and the first migration: the `users` table
- [ ] 4. Sign up: the `auth.v1` API in Protobuf, with passwords stored as argon2id hashes
- [ ] 5. Log in: create a session, send its token in an HttpOnly cookie, store only the token's hash
- [ ] 6. Devices: each login is a device. List your devices, log out, log out another device.
- [ ] 7. The web app: sign-up, login and "logged in as …" pages
- [ ] 8. Security basics and tests: rate limiting on login, integration tests with Testcontainers

The steps can change as we learn. Any change is written here.

## Decisions

- [0012](../decisions/0012-use-pgx-for-database-access.md): sqlc joins pgx in this milestone (decided in milestone 1)
- [0021](../decisions/0021-log-in-with-a-username-and-open-sign-up.md): log in with a username, open sign-up
- [0022](../decisions/0022-keep-sessions-in-httponly-cookies.md): sessions are random tokens in an HttpOnly cookie
- [0023](../decisions/0023-use-testcontainers-for-integration-tests.md): Testcontainers for integration tests
- TanStack Query: not yet. Plain Connect calls are enough for a few pages ([brief](../PROJECT_BRIEF.md#future-considerations)).

## Concepts learned

Filled in as we go, with links to `docs/learn/`.

## Done when

`task dev` starts everything. In the browser you can sign up, log in, reload the page and still be logged in, see your devices, and log out. Tests and CI pass.
