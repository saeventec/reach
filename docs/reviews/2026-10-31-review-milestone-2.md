# Review: Milestone 2, accounts

- **Date:** 2026-10-31
- **Commit or change:** the branch `milestone-2-accounts`, reviewed by the owner while it was built

## Feedback

### 1. Don't wait for approval before committing

- **Question or comment:** commit each step when it is finished, without waiting, and explain the changes in the summary.
- **Answer or fix:** added to the workflow in [AGENTS.md](../../AGENTS.md) (rule 6), the brief and the milestone file. Big design decisions are still asked first.

### 2. A task to stop the system

- **Question or comment:** add a task to stop the project after `task dev`.
- **Answer or fix:** `task stop` runs `docker compose down`: it stops and removes the containers and the network, and keeps the database's volume.

### 3. Connect to the database from tools outside Docker

- **Question or comment:** how can a VS Code extension reach the database?
- **Answer or fix:** Postgres is published on `127.0.0.1:5432` only, so other devices on the network can't reach it. The Microsoft PostgreSQL extension is recommended. See [postgresql.md](../learn/postgresql.md#connect-from-your-computer).

### 4. Where is the session expiry checked?

- **Question or comment:** the owner couldn't find the expiry check in the Go code.
- **Answer or fix:** it is in the SQL (`expires_at > now()` in `RenewSession` and `ListSessions`), so no Go code can forget it. A comment in `authenticate` now points there, and the integration tests check that an expired session is refused.

## What I learned

- A rule can live in the database instead of the code. Then it holds for every caller, but it is harder to find, so point to it with a comment.
- A test is only proven when it fails for the bug it is meant to catch.
