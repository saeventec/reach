# 0001. Use a monorepo

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

Reach has many parts: a web app, several services, shared code, the API contract, infrastructure and docs. We need to decide where they live.

## Decision

Keep everything in one GitHub repository, `reach`, organized in folders by purpose (`apps/`, `services/`, `packages/`, `proto/`, `docs/`, ...).

## Why

- One place to clone, search and read: easier to learn the whole system.
- A change that touches the API, a service and the web app is one commit, reviewed together.
- Agents get the full context from one repo.

## Alternatives considered

- **One repo per service (polyrepo):** clear ownership, but changes across repos are slow and hard to keep in sync.
- **Two repos (frontend and backend):** simpler than many repos, but the shared API contract would be split.

## Consequences

The repo grows large over time, so folders must stay well organized. CI must run only the checks that matter for what changed (later).
