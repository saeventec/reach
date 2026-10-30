# 0011. Start with a small hello service

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Milestone 1 needs one Go service that reads a greeting from Postgres, so we can see a request travel through the whole system. The planned services are `auth`, `chat` and `realtime`, but none of them is about greetings.

## Decision

Create a new, small service called `hello` (in `services/hello/`) just for milestone 1.

## Why

- The name is honest: no code pretends to be `auth` or `chat`.
- It is safe to break and experiment with while learning.
- Its base (config, health check, Dockerfile, database connection) becomes the template for `auth` and `chat`.

## Alternatives considered

- **Start `auth` now:** a real service, but the greeting code would be thrown away in milestone 2.
- **Start `chat` now:** same idea, but chat is only needed in milestone 3.

## Consequences

When `auth` starts (milestone 2), we copy the useful parts and then delete `hello` or keep it as a small example. That choice is made then.
