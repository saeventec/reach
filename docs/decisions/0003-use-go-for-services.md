# 0003. Use Go for services

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The backend services need a language that is fast, good with many network connections, and possible to learn fully.

## Decision

Write all backend services in Go.

## Why

- Small language: it can be learned completely.
- Great at many connections at once (needed for chat and WebSocket).
- Compiles to one small file, which makes tiny containers.
- A strong standard library (HTTP, JSON, logging) means few dependencies.

## Alternatives considered

- **Node.js / TypeScript:** same language as the web app, but a weaker fit for long-lived, heavy services.
- **Rust:** very fast and safe, but much harder to learn.
- **Kotlin or C#:** powerful, but heavier runtimes and bigger ecosystems to learn.

## Consequences

The project uses two languages (Go and TypeScript). Protobuf (ADR 0005) keeps them in agreement.
