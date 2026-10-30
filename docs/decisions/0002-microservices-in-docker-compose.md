# 0002. Microservices in Docker Compose

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The owner wants to learn real microservices, but the whole system must start on one computer with one command.

## Decision

Each service is a separate Go program in its own container. Docker Compose starts all of them, plus Postgres, NATS and the gateway, with one command. We start with a few meaningful services (`auth`, `chat`, `realtime`) and split further only with a clear reason.

## Why

- Teaches service boundaries, events and containers in practice.
- Compose runs everything locally with one file and one command.
- The same containers move to Kubernetes later (one Compose service becomes one deployment).

## Alternatives considered

- **One program (monolith):** simpler, but teaches less about distributed systems.
- **Modular monolith first, split later:** a good real-world path, but delays the learning goal.
- **Kubernetes from day one:** too much complexity before there is anything to run.

## Consequences

More moving parts than a monolith: services must be stateless, configured by environment variables, and have health checks. Each service owns its own data.
