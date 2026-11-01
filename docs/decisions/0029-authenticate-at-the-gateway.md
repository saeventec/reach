# 0029. Authenticate at the gateway with forward_auth

- **Status:** Accepted (the owner's choice)
- **Date:** 2026-11-01

## Context

Until now only `auth` reads the session cookie. The new `chat` and `realtime` services (milestone 3) also need to know who is calling. See the [design doc](../design/m03-real-time-chat.md#1-who-is-asking-authentication-at-the-gateway).

## Decision

Caddy's `forward_auth` sends a copy of each `chat` and `realtime` request to a small endpoint in `auth`. A valid session gets `200` with an `X-User-Id` header, which Caddy adds to the real request. Anything else gets `401`, and the request never reaches the service.

## Why

- Sessions are checked in one place. `chat` and `realtime` read one header and never see the cookie.
- Logging out a device works instantly, because every request is checked against the database.
- The same pattern big companies use in their API gateways (for example Envoy's "external authorization").

## Alternatives considered

- **Each service calls `auth` itself:** no trusted header, but every service repeats the same code.
- **Signed tokens (JWT) checked by each service:** no call per request, but logout isn't instant and keys must be managed. A good later topic.

## Consequences

- Services trust `X-User-Id`, so only the gateway may reach them. Caddy removes any `X-User-Id` the browser sends, the services have no published ports, and a test sends a forged header and expects a refusal. (Caddy's docs don't say if `forward_auth` removes it by itself, so we remove it ourselves.)
- One extra call to `auth` and one database read per request. Fine at our scale; caching is a later topic.
- Found while building: Caddy runs `forward_auth` before `request_header` unless both are inside a `route` block. Without it, the real `X-User-Id` was removed too. The snippet in `gateway/Caddyfile` uses `route`; keep it when editing.
