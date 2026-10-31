# 0016. Run database migrations in a separate container

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Each service has goose migrations: versioned SQL files that build its database step by step. They must run before the service starts, automatically, as part of `task dev`.

## Decision

Each service gets a small migrate container (for example `hello-migrate`). It runs `goose up` once and exits. Compose starts it after Postgres is healthy, and starts the service only after the migrations finished successfully.

## Why

- Migrations stay separate from the app code: the service never changes its own database structure.
- It works like a Kubernetes Job later, so the move to Kubernetes needs no redesign.
- If many copies of a service start at once, the migrations still run only once.

## Alternatives considered

- **Inside the service, at startup (goose as a Go library):** one container fewer, but every copy of the service tries to migrate, and the app code mixes with database setup.
- **By hand (`task migrate`):** very visible, but a manual step, so `task dev` alone would not give a working system.

## Consequences

Each service has two images: the service and its migrations. goose has no official image, so we compile it in the migrate Dockerfile. The first build takes a few minutes; later builds use Docker's cache.
