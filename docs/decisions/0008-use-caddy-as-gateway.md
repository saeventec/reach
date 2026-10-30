# 0008. Use Caddy as the gateway

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The browser should talk to one address only, which routes each request to the right service.

## Decision

Use Caddy as the single entry point (gateway), configured with a `Caddyfile` in `gateway/`.

## Why

- Very short, readable configuration.
- Automatic HTTPS later, when Reach runs on a server.
- Supports WebSocket out of the box.

## Alternatives considered

- **Traefik:** discovers containers automatically, but its configuration is harder to read.
- **Nginx:** very popular and fast, but more manual configuration (and HTTPS setup).
- **Envoy:** powerful, but complex. Built for very large systems.

## Consequences

All routing rules live in one file, which must be kept in sync when services are added.
