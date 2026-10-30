# 0015. Use Alpine as the runtime image for Go services

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

A Go service compiles to one program file. Its container image needs a small base image to run that file in. Docker also needs a way to call `/healthz` from inside the container.

## Decision

Build with `golang:<version>-alpine` and run on `alpine:<version>`, in a two-stage Dockerfile. Pin exact versions.

## Why

- Alpine is a very small Linux (the final hello image is about 28 MB).
- It has a shell and `wget`, so the Compose health check needs no extra code, and you can look inside a running container while learning (`docker compose exec hello sh`).
- It is widely used and well documented.

## Alternatives considered

- **Distroless (`gcr.io/distroless/static`):** smaller and safer (no shell), but the health check then needs extra code in every service.
- **scratch (an empty image):** the smallest, but has nothing at all: no shell, no certificates, no time zones.
- **Debian slim:** bigger, and has more than we need.

## Consequences

Base image versions are updated by hand for now (later with Renovate). We can move to distroless later, when we want a smaller attack surface; it only changes the Dockerfile and the health check.
