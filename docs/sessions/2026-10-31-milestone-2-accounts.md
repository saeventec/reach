# Session: Milestone 2, accounts

- **Date:** 2026-10-31
- **Milestone:** m02

## Goal

Build milestone 2: sign up, log in, sessions and devices, in a new `auth` service and the web app, with security basics and integration tests. Work on the branch `milestone-2-accounts` and merge it through a pull request.

## What was done

All 8 steps of the [milestone plan](../milestones/m02-accounts.md):

1. The plan and ADRs 0021 to 0023.
2. The `auth` service skeleton (copied from `hello`). Each service now logs in to Postgres as its own user, which owns only its own database. The `postgres-setup` container creates them on every start, safely repeatable, and moved the existing `hello` tables to the `hello` user.
3. sqlc and the `users` table, with UUIDv7 ids and the username rules as `CHECK` constraints.
4. Sign-up: username and password rules, argon2id hashing (OWASP settings, PHC format), the `auth.v1.AccountService` API, and the Caddy route.
5. Login: the `sessions` table, password checking, random tokens stored only as SHA-256 hashes, an `HttpOnly; SameSite=Strict` cookie (`Secure` unless `COOKIE_SECURE=false`), renewed on every use for 30 days. A wrong username and a wrong password get the same answer in the same time.
6. Devices: each session remembers its browser and last use. `ListDevices`, `LogOut` and `LogOutDevice` (only your own devices).
7. The web app: React Router with `/login`, `/signup` and `/account`, the forms, your devices, and log out. The greeting stays at the bottom.
8. A 16 KiB limit on API messages, rate limiting (per IP address and per username, in memory), and integration tests against a real Postgres with Testcontainers.

Asked for by the owner along the way:

- `task stop` stops the system and removes its containers (the data is kept).
- Postgres is published on `127.0.0.1:5432`, so database tools outside Docker can connect. The VS Code PostgreSQL extension is recommended.
- A comment in `authenticate` that points to where the session expiry is checked (in the SQL).
- A new workflow rule: commit each finished step right away; the owner reviews afterwards ([AGENTS.md](../../AGENTS.md)).

A final review fixed stale doc lines, moved sqlc, Testcontainers, argon2id and rate limiting into the brief's main tech table, and split `LogIn` into smaller functions.

## What was explained

- Database: [one user per service, idempotent setup](../learn/postgresql.md#one-user-per-service), [connecting from your computer](../learn/postgresql.md#connect-from-your-computer), [UUIDv7, constraints, regular expressions](../learn/postgresql.md#ids-and-constraints), [sqlc](../learn/sqlc.md), foreign keys and indexes ([sessions.md](../learn/sessions.md))
- Security: [password hashing, salts, NFC, timing attacks](../learn/password-hashing.md), [sessions, cookies, CSRF and SameSite](../learn/sessions.md), [rate limiting](../learn/rate-limiting.md), [message size limits](../learn/connect-rpc.md#a-limit-on-message-size), [settings with a safe default](../learn/environment-variables.md#settings-with-a-safe-default)
- API: [lists, Timestamps and default values](../learn/protobuf.md#lists-dates-and-default-values), [one API per service in buf](../learn/protobuf.md#one-api-per-service)
- Frontend: [React Router](../learn/react-router.md), [forms](../learn/react.md#forms), [lists and keys](../learn/react.md#lists-and-keys)
- Testing: [table-driven tests](../learn/testing.md#table-driven-tests), [integration tests](../learn/testing.md#integration-tests), and breaking code on purpose to prove a test can fail
- Tooling: [why the Lefthook rc path needs `./`](../learn/lefthook.md)

## Decisions

All chosen or confirmed by the owner:

- [0021](../decisions/0021-log-in-with-a-username-and-open-sign-up.md) username login, open sign-up; lowercase usernames, 3 to 32 characters, `a-z 0-9 . _ -`
- [0022](../decisions/0022-keep-sessions-in-httponly-cookies.md) random tokens in an HttpOnly cookie; 30 days renewed on use, `SameSite=Strict`, `COOKIE_SECURE`
- [0023](../decisions/0023-use-testcontainers-for-integration-tests.md) Testcontainers; [0024](../decisions/0024-create-databases-with-a-setup-container.md) a setup container for databases and users
- [0025](../decisions/0025-password-rules-and-argon2id.md) passwords of 8 to 128 characters, not the username, no "reach" (below NIST's advice, to revisit before a real server)
- [0026](../decisions/0026-use-react-router.md) React Router; [0027](../decisions/0027-rate-limit-logins-in-memory.md) rate limits in memory, by IP and username
- Shared Go code is copied for now (until a third service). TanStack Query, password change, account deletion and a maximum session age: not now.

## Problems solved

- The pre-commit hook failed on WSL2: dash, its `sh`, doesn't look in the current folder for `. .lefthookrc`. Fixed with `./.lefthookrc`.
- buf's `clean` would have deleted sqlc's generated code, which lived in the same `gen/` folder. buf now writes to `gen/proto/`.
- argon2 crashes on 0 passes or threads; a damaged stored hash now gives an error instead.
- Caddy replaces a faked `X-Forwarded-For`, which was tested with 21 fake addresses.
- Mistakes by the agent, all caught before they mattered: an unused constant and an unchecked `Close()` (caught by the linter), a test that passed for the wrong reason, a wrong comment about how browsers match cookies, a browser check against a stale container (started without `--watch`), and a cleanup `rm` with an unquoted variable, stopped by a safety check.

## Open questions

Known gaps, kept for later on purpose (written in the ADRs and notes):

- Password rules below NIST's advice (ADR 0025) and real client addresses behind Docker (ADR 0027): before a real server, milestone 10.
- Rate-limit counts are per copy of the service (ADR 0027): milestone 11.
- Expired sessions are never deleted; the built app must be served with an `index.html` fallback; the goose Dockerfile and service base are copied twice; the first auth image build is slower because of test libraries.
- Not yet checked on the Mac: `task setup`, three new `.env` lines, and login in Safari.

## Next step

After the pull request is merged: check milestone 2 on the Mac, then plan [milestone 3: real-time chat](../PROJECT_BRIEF.md#milestones) (1:1 and groups, WebSocket, NATS, message ordering). Decide then: the `chat` and `realtime` services, sharing Go code in `packages/go/`, and Playwright for end-to-end tests.
