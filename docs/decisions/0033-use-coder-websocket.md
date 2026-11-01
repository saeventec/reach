# 0033. Use coder/websocket in the realtime service

- **Status:** Accepted (proposed by the agent)
- **Date:** 2026-11-01

## Context

The brief chose WebSocket for live delivery. Go's standard library has no WebSocket server, and `golang.org/x/net/websocket` is marked as lacking features and points to other packages.

## Decision

`realtime` uses [github.com/coder/websocket](https://github.com/coder/websocket).

## Why

- Maintained (by Coder since 2024), small, with few dependencies.
- `context` support on reads and writes, like the rest of our Go code.
- Safe to write from several goroutines at once.
- Refuses connections from other websites by default (it checks the `Origin` header), which stops cross-site WebSocket hijacking.

## Alternatives considered

- **gorilla/websocket:** the most famous one, stable, but little activity and no `context` support.
- **golang.org/x/net/websocket:** its own docs point to the two above.
- **Write it ourselves on `net/http`:** a good exercise, but a lot of protocol detail to get right.

## Consequences

- One new Go dependency in `realtime`.
