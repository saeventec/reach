# Builds a small image that applies the hello service's database migrations
# with goose, then exits. Compose runs it before the hello service starts.
# Learn more: docs/learn/database-migrations.md

# Stage 1: compile goose from its source code (goose has no official image).
FROM golang:1.27.3-alpine3.24 AS build

RUN CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.29.0

# Stage 2: a small image with only goose and our SQL files.
FROM alpine:3.24.3

COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY *.sql /migrations/

USER nobody

# goose reads its settings from environment variables. The folder with the
# SQL files is fixed here. The database address comes from compose.yaml.
ENV GOOSE_MIGRATION_DIR=/migrations

# Apply every migration that has not been applied yet.
ENTRYPOINT ["goose", "up"]
