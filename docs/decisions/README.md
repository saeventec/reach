# Decisions

Decision records (ADRs, "Architecture Decision Records"): one decision per file, with what we chose, why, and the alternatives. Read one before changing something it decided.

## Index

| ADR | Decision | Status |
| --- | --- | --- |
| 0001 | [Use a monorepo](0001-use-a-monorepo.md) | Accepted |
| 0002 | [Microservices in Docker Compose](0002-microservices-in-docker-compose.md) | Accepted |
| 0003 | [Use Go for services](0003-use-go-for-services.md) | Accepted |
| 0004 | [React, TypeScript and Vite for the web app](0004-react-typescript-vite-for-web.md) | Accepted |
| 0005 | [Protobuf and Connect for the API](0005-protobuf-and-connect-for-the-api.md) | Accepted |
| 0006 | [Use PostgreSQL](0006-use-postgresql.md) | Accepted |
| 0007 | [Use NATS for service messaging](0007-use-nats-for-service-messaging.md) | Accepted |
| 0008 | [Use Caddy as the gateway](0008-use-caddy-as-gateway.md) | Accepted |
| 0009 | [Use mise and Task for tooling](0009-use-mise-and-task-for-tooling.md) | Accepted |
| 0010 | [Docs as code in docs/](0010-docs-as-code.md) | Accepted |
| 0011 | [Start with a small hello service](0011-start-with-a-hello-service.md) | Accepted |
| 0012 | [Use pgx for database access, sqlc later](0012-use-pgx-for-database-access.md) | Accepted |
| 0013 | [Use Tailwind CSS for styling](0013-use-tailwind-for-styling.md) | Accepted |
| 0014 | [Use Lefthook for checks before each commit](0014-use-lefthook-for-commit-checks.md) | Accepted |
| 0015 | [Use Alpine as the runtime image for Go services](0015-use-alpine-as-the-runtime-image.md) | Accepted |
| 0016 | [Run database migrations in a separate container](0016-run-migrations-in-a-separate-container.md) | Accepted |
| 0017 | [Use buf with local generators, and commit the generated code](0017-use-buf-with-local-generators.md) | Accepted |
| 0018 | [Run the web app's development server in Compose, with Compose Watch](0018-run-the-web-dev-server-in-compose.md) | Accepted |
| 0019 | [Use Oxlint and Prettier for the web app](0019-use-oxlint-and-prettier-for-the-web-app.md) | Accepted |
| 0020 | [CI with GitHub Actions, using mise and Task](0020-ci-with-github-actions-mise-and-task.md) | Accepted |
| 0021 | [Log in with a username, open sign-up](0021-log-in-with-a-username-and-open-sign-up.md) | Accepted |
| 0022 | [Keep sessions in HttpOnly cookies](0022-keep-sessions-in-httponly-cookies.md) | Accepted |
| 0023 | [Use Testcontainers for integration tests](0023-use-testcontainers-for-integration-tests.md) | Accepted |
| 0024 | [Create each service's database and user with a setup container](0024-create-databases-with-a-setup-container.md) | Accepted |
| 0025 | [Password rules, and hashing with argon2id](0025-password-rules-and-argon2id.md) | Accepted |
| 0026 | [Use React Router for the web app's pages](0026-use-react-router.md) | Accepted |
| 0027 | [Rate limit logins and sign-ups in memory](0027-rate-limit-logins-in-memory.md) | Accepted |
| 0028 | [Reach is a learning showcase: real concepts, tiny scale](0028-reach-is-a-learning-showcase.md) | Accepted |
| 0029 | [Authenticate at the gateway with forward_auth](0029-authenticate-at-the-gateway.md) | Accepted |
| 0030 | [Hexagonal architecture for the chat service](0030-hexagonal-architecture-for-chat.md) | Accepted |
| 0031 | [Publish chat events with a transactional outbox](0031-transactional-outbox-for-chat-events.md) | Accepted |
| 0032 | [Order messages by a sequence number per conversation](0032-order-messages-by-sequence-per-conversation.md) | Accepted |
| 0033 | [Use coder/websocket in the realtime service](0033-use-coder-websocket.md) | Accepted |
| 0034 | [Use Playwright for end-to-end tests](0034-use-playwright-for-end-to-end-tests.md) | Accepted |
| 0035 | [Share Go code in packages/go](0035-share-go-code-in-packages-go.md) | Accepted |
| 0036 | [Readable code and token-efficient docs](0036-readable-code-and-token-efficient-docs.md) | Accepted |

## How to add one

1. Copy `_template.md`, name it with the next number: `0037-short-title.md`.
2. Fill it in. Keep it short.
3. Add it to the index.

To change a decision, write a new record and set the old one's status to "Replaced by 00XX". Never delete old records: they are history.
