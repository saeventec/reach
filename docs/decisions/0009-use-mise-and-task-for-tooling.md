# 0009. Use mise and Task for tooling

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The owner works on a Mac and on Windows (WSL2), and CI runs on Linux. All three must use the same tool versions and the same commands.

## Decision

Use mise to install exact tool versions from `.mise.toml`, and Task to run project commands from `Taskfile.yml`.

## Why

- mise: one file pins Go, Node.js and Task versions for every machine and CI.
- Task: one simple YAML file of commands that works the same on every system.
- A fresh machine only needs Git, Docker and mise.

## Alternatives considered

- **asdf or nvm (instead of mise):** asdf is slower and older; nvm only handles Node.js.
- **Make or just (instead of Task):** Make is everywhere but has a tricky syntax; just is nice but less common.

## Consequences

Every developer installs mise and activates it in their shell once. New tools are added to `.mise.toml`, new commands to `Taskfile.yml`.
