# Session: Milestone 3, real-time chat

- **Date:** 2026-11-01
- **Milestone:** m03

## Goal

Build [milestone 3](../milestones/m03-real-time-chat.md): logged-in people start 1:1 chats and groups, send messages that arrive live, and see "delivered" and "read". Two new services (`chat` and `realtime`), connected by NATS. Plan: the [design doc](../design/m03-real-time-chat.md).

## What was done

All 14 steps, in 25 commits on the branch `milestone-3-chat` (not pushed yet):

- **Plan:** the design doc, the milestone file, and ADRs 0029 to 0035.
- **Shared Go code** in `packages/go` (config, health check, database, clean shutdown, and later the NATS address and chat's event subjects). All generated Go code moved to `packages/go/gen`.
- **NATS with JetStream** in Compose.
- **Authentication at the gateway:** a session check in `auth`, and Caddy's `forward_auth`, which removes forged `X-User-Id` headers.
- **`chat`**, in hexagonal style:
  - conversations (one 1:1 per pair, groups of 2 to 50);
  - messages with a sequence number per conversation and app-made UUIDv7 ids (a retry is stored once);
  - delivered and read watermarks;
  - usernames from a new internal `auth` API.
- **Events through a transactional outbox:** `chat.message.sent`, `chat.receipt.updated` and `chat.conversation.created`, saved in the same transaction as the change and published to JetStream by a relay.
- **`realtime`:** WebSockets with coder/websocket, a mutex-protected hub, slow browsers dropped, "going away" on a stop.
- **The web app:**
  - the conversation list and the new chat and new group forms;
  - the conversation page, with older messages and sending with retries;
  - live updates with reconnecting and catching up (`ConversationSync`);
  - ✓, ✓✓ and "Read".
- **Tests:**
  - unit tests and Postgres and NATS integration tests, including one of the store, the relay and NATS together;
  - the race detector (`task test:race`);
  - six Playwright end-to-end tests (`task test:e2e`), including forged headers through the real gateway.
- **A fix:** generated TypeScript enums broke `task build` from step 6 to step 11 (`erasable_syntax=true`).

## What was explained

- Distributed systems:
  - [the transactional outbox, at-least-once delivery, event notification versus event-carried state](../learn/transactional-outbox.md);
  - [ordering, idempotency, watermarks, catching up](../learn/message-ordering.md);
  - [NATS and JetStream, acks, dropping copies with `Nats-Msg-Id`](../learn/nats.md).
- Architecture:
  - [hexagonal architecture](../learn/hexagonal-architecture.md);
  - [authentication at the gateway](../learn/caddy.md#checking-the-session-at-the-gateway);
  - [sharing Go code](../learn/go.md#sharing-code-between-modules).
- Databases: [transactions, row locks, `ON CONFLICT`, identity columns, partial indexes](../learn/postgresql.md).
- Concurrency:
  - [WebSockets, the hub, mutexes, the race detector, stopping cleanly](../learn/websocket.md);
  - [exponential backoff](../learn/websocket.md#in-the-browser).
- Frontend:
  - [custom hooks, `useRef`, reloading with an effect](../learn/react.md#custom-hooks);
  - [URL parameters, `key`, `NavLink`](../learn/react-router.md);
  - [Protobuf `oneof`, `int64` in JSON, zero values, enums as objects](../learn/protobuf.md).
- Testing:
  - [Playwright, the test pyramid, locators, flaky tests](../learn/playwright.md);
  - "mutation checks": breaking the code on purpose to prove a test can fail. This found one useless test (step 7) and confirmed the others.

## Decisions

- **The owner's choices:**
  - [0029](../decisions/0029-authenticate-at-the-gateway.md): authenticate at the gateway;
  - [0030](../decisions/0030-hexagonal-architecture-for-chat.md): hexagonal `chat`;
  - [0031](../decisions/0031-transactional-outbox-for-chat-events.md): the transactional outbox;
  - [0034](../decisions/0034-use-playwright-for-end-to-end-tests.md): Playwright.
- **The agent's proposals:**
  - [0032](../decisions/0032-order-messages-by-sequence-per-conversation.md): sequence numbers and watermarks;
  - [0033](../decisions/0033-use-coder-websocket.md): coder/websocket;
  - [0035](../decisions/0035-share-go-code-in-packages-go.md): `packages/go`.
- **Made while building, written in the design doc's "Added while building" notes:**
  - a 4,000-character message limit;
  - JetStream message ids, and a 24-hour stream;
  - no `SKIP LOCKED` yet;
  - the race detector in its own task, because it needs a C compiler;
  - watermarks capped at the newest message;
  - `chat.conversation.created` carries only ids;
  - e2e tests in `apps/web/e2e`, as a separate Compose project.
- **The owner's answers this session:**
  - the race detector stays out of `task test`;
  - don't push yet;
  - there are no stale reminders about `.env`. The agent wrongly repeated one; it now checks first.

## Open questions

- **CI:** the new CI steps (the race detector, Playwright) have only run locally. The first push shows whether they pass on GitHub.
- **auth's rate limit and more e2e tests:** 20 sign-ups and logins per address; today's tests use 16. When more are needed:
  - (A, recommended) a limit setting that only the test system raises;
  - (B) reusing logins (Playwright's storage state).
- **Suggestions, not done:**
  - a request size limit for `hello` (it has none, unlike `auth` and `chat`);
  - `chat` in `task test:race` too (it passes today);
  - deleting old outbox rows;
  - a `GetConversation` call, instead of the page loading the whole list;
  - one shared WebSocket for all pages;
  - scrolling to the newest message;
  - rate limiting message sending.
- **Later by design:**
  - the sender's own watermarks (for unread counts);
  - per-device delivered;
  - end-to-end encryption (milestone 8);
  - several copies of `realtime` and `chat` (a good experiment).
- `erasable_syntax=true` in protoc-gen-es is marked "experimental". Watch it when updating.

## Next step

The owner reviews the commits on `milestone-3-chat`, and saves the feedback in `docs/reviews/`. Then push, open a pull request, and check that CI passes before merging. Try it by hand: `task dev`, two browser windows, alice and bob. After that: the design doc for milestone 4, the first release.
