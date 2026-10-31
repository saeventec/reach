# Builds a small image that applies the auth service's database migrations
# with goose, then exits. Compose runs it before the auth service starts.
# Learn more: docs/learn/database-migrations.md

# Stage 1: compile goose from its source code (goose has no official image).
FROM golang:1.27.3-alpine3.24 AS build

# goose includes drivers for many databases. Each "no_..." build tag leaves
# one out. We only keep Postgres, so the image is smaller and builds faster.
RUN CGO_ENABLED=0 go install \
    -tags="no_clickhouse no_libsql no_mssql no_mysql no_sqlite3 no_vertica no_ydb" \
    github.com/pressly/goose/v3/cmd/goose@v3.29.0

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
