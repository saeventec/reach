# 0014. Use Lefthook for checks before each commit

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Format and lint mistakes should be caught on your own computer, before they reach GitHub and fail in CI.

## Decision

Use Lefthook to run the format and lint checks automatically before each `git commit`. It is added in step 7 of milestone 1, together with the linters.

## Why

- Mistakes are found in seconds, not after a push.
- One small config file, and it works the same on Mac, Windows (WSL2) and Linux.
- Its version can be pinned in `.mise.toml` like our other tools.

## Alternatives considered

- **pre-commit:** popular, but needs Python.
- **Husky:** popular, but made for JavaScript projects and needs npm.
- **No hook:** CI catches the same mistakes, only later.

## Consequences

Each developer runs a one-time install command (added to `task setup`). CI still runs every check, because a hook can be skipped.
