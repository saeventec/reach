# 0027. Rate limit logins and sign-ups in memory

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Login and sign-up must not allow unlimited tries: an attacker could guess passwords (brute force) or create accounts in bulk. We need to decide where the counts are kept and what they count by.

## Decision

The auth service counts tries with token buckets from `golang.org/x/time/rate`, kept **in memory** (the owner's choice), by **IP address and by username** (the owner's choice):

- Per IP address: 20 tries at once, then 20 per minute, for every login and sign-up.
- Per username: 5 failed logins at once, then 5 per 15 minutes. Successful logins don't count.
- Over the limit: `resource_exhausted`, "too many attempts, please try again later".

The client's address comes from the `X-Forwarded-For` header set by Caddy, which replaces any value sent by the client.

## Why

- In memory is the simplest: no new container, no extra database write per login.
- The counts are short-lived and not important data, so losing them on a restart is fine, and the service stays effectively stateless.
- Two limits cover two attacks: many accounts from one computer, and one account from many computers.
- A slow-down instead of a lock: nobody can lock someone else out for long.

## Alternatives considered

- **Counts in Postgres:** shared and kept on restart, but a write on every login, and a cleanup job.
- **Counts in Valkey:** shared and fast, but one more container to learn and run now.
- **A Caddy rate-limit plugin:** needs a custom Caddy build.
- **Only per IP, or only per username:** each one leaves one of the two attacks open.

## Consequences

- With several copies of the service (milestone 11), each copy counts on its own, so the real limit grows with the number of copies. Then move the counts to a shared store (Valkey).
- **On a real server (milestone 10):** check that Docker passes the real client addresses through to Caddy. Locally all requests come from Docker's gateway address (`172.18.0.1`), so everyone shares one address limit.
- The numbers are constants in [services/auth/rate_limit.go](../../services/auth/rate_limit.go), easy to change.
