# 0035. Share Go code in packages/go

- **Status:** Accepted (proposed by the agent, following milestone 2's plan)
- **Date:** 2026-11-01

## Context

`hello` and `auth` each have their own copy of the same code: reading settings, the health check, connecting to Postgres, and stopping cleanly. Milestone 2 decided to share it when a third service arrives; `chat` is the third, `realtime` the fourth.

## Decision

The common code moves to one small Go module, `packages/go`. Each service points to it with a `replace` line in its `go.mod`. Each service's image is then built with the whole repo as its build context, so the Dockerfile can copy `packages/go`.

## Why

- Four copies would slowly drift apart.
- `replace` works the same locally, in CI and in Docker.

## Alternatives considered

- **A `go.work` file (Go workspaces):** nicer for editors, but every Docker build then needs every module listed in it.
- **Keep copying:** fine for two copies, not four.
- **Publish it as its own versioned module:** the "real" way across repos, but too much ceremony inside one repo.

## Consequences

- A change in `packages/go` can affect every service, so it stays small and well tested.
- A `.dockerignore` at the repo root keeps the build context small.
