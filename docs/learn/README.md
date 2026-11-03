# Learning notes

The owner's textbook: one short note per concept or tool, written the first time it appears. Code comments link here instead of repeating the explanation.

## Index

**Repository and tooling**
- [Git](git.md) · [Monorepo](monorepo.md) · [Markdown](markdown.md)
- [mise](mise.md) · [Task](task.md) · [npm and package.json](npm.md)
- [.gitignore](gitignore.md) · [.gitattributes](gitattributes.md) · [.editorconfig](editorconfig.md)
- [Linting and formatting](linting.md) · [Git hooks and Lefthook](lefthook.md) · [CI with GitHub Actions](ci.md)

**Go services**
- [Go](go.md) · [HTTP and web servers](http.md) · [Environment variables](environment-variables.md)
- [Logging](logging.md) · [Graceful shutdown](graceful-shutdown.md)

**Containers and gateway**
- [Docker](docker.md) · [Docker Compose](docker-compose.md) · [Caddy and the gateway](caddy.md)

**Databases**
- [PostgreSQL](postgresql.md) · [Database migrations (goose)](database-migrations.md)
- [pgx: talking to Postgres from Go](pgx.md) · [sqlc: Go code from SQL queries](sqlc.md)

**APIs**
- [Protobuf, buf and code generation](protobuf.md) · [APIs and Connect RPC](connect-rpc.md)

**Security**
- [Password hashing (argon2id)](password-hashing.md) · [Sessions and tokens](sessions.md) · [Rate limiting](rate-limiting.md)

**Messaging and real time**
- [NATS and JetStream](nats.md) · [WebSocket and the realtime service](websocket.md)
- [Message ordering and idempotency](message-ordering.md) · [The transactional outbox](transactional-outbox.md)

**Architecture**
- [Hexagonal architecture](hexagonal-architecture.md)

**Web app**
- [React](react.md) · [React Router](react-router.md) · [Vite](vite.md)
- [TypeScript](typescript.md) · [Tailwind CSS](tailwind.md)

**Testing**
- [Testing](testing.md) · [Playwright: end-to-end tests](playwright.md)

## How to add one

1. Copy `_template.md`, name it after the concept in lowercase: `websocket.md`.
2. Keep it short: what it is, why we use it, a tiny example, the alternatives.
3. Explain each concept in one note only; other notes link to it.
4. Add it to the right group in the index.
