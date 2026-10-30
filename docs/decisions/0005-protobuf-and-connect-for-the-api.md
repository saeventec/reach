# 0005. Protobuf and Connect for the API

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The web app and the Go services must agree on every request and response. Later, native apps and offline transports need the same message format.

## Decision

Define the API and events once in Protobuf files (`proto/`). Generate Go and TypeScript code from them. Use Connect RPC to call the API from the browser.

## Why

- One source of truth: front and back can never disagree.
- Connect also speaks plain JSON over HTTP, so requests are easy to read while debugging.
- The same format is reused for NATS events and, later, offline transports.

## Alternatives considered

- **REST + OpenAPI:** very common, but the contract is easier to drift from the code.
- **gRPC:** same idea, but browsers can't call it directly without a proxy.
- **GraphQL:** flexible queries, but more complex on the server.

## Consequences

We need a code generation step (with buf, decided in milestone 1). Changing messages must follow Protobuf rules so older clients keep working.
