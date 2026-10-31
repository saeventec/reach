# 0020. CI with GitHub Actions, using mise and Task

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

Every change needs an automatic, independent check: lint, tests, generated code, API compatibility and builds. The brief chose GitHub Actions. We still need to decide how the workflow gets its tools and commands.

## Decision

One workflow, `.github/workflows/ci.yml`, with one job. It installs the tools from `.mise.toml` with the official `jdx/mise-action`, and then only runs our Task commands (`task setup`, `task lint`, `task generate`, `task test`, `task build`), plus `buf breaking`. Third-party actions are pinned to exact commits.

## Why

- Same tool versions and same commands as on your computer: a CI failure can always be reproduced locally.
- The workflow stays short. New checks go into the Taskfile, and CI picks them up.
- Pinning actions to commits and giving the workflow read-only rights are GitHub's own security recommendations.

## Alternatives considered

- **One setup action per language (setup-go, setup-node, ...):** common, but versions could drift from `.mise.toml`.
- **Several parallel jobs (lint, test, build):** faster for big projects; one job is simpler and fast enough now.
- **Self-hosted CI (Woodpecker, Forgejo Actions):** fully self-hosted, but GitHub is an allowed service (Golden rule 9) and needs no maintenance.

## Consequences

Pinned action versions must be updated on purpose (later automatically with Renovate). If CI gets slow, we can split the job.
