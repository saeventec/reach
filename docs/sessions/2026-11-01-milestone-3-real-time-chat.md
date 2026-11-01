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
- **At the end, the open suggestions, applied** (the owner asked):
  - `hello` caps requests at 1 KiB, like `auth` and `chat`;
  - `task test:race` covers `chat` too;
  - a limit on sending: 30 messages at once per user, then one per second;
  - auth's per-address limit is a setting, `RATE_LIMIT_BURST_PER_IP`, which only the e2e test system raises (the suite now passes 5 times in one run);
  - the token buckets moved to `packages/go/ratelimit`;
  - published outbox events are deleted after 7 days;
  - `GetConversation`, so the page doesn't load the whole list;
  - scrolling to the newest message.
  - Not done, because not needed: one shared WebSocket. Only one page is open at a time.
- **A review of the whole branch** (the code-review skill) found 10 points. Fixed:
  - stale "Read" marks after a reconnect;
  - a failed mark never tried again;
  - your own messages marked read;
  - delivered marked twice on the list;
  - an empty message hiding an error in the retry path;
  - rows that could land on the wrong conversation;
  - a duplicated helper, a misleading comment, and a line doing too much.

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

- **For the owner to decide:** every chat call and WebSocket open goes through the gateway's session check, which renews the session (a database write in `auth`). It also keeps the device list's "last used" exact. Fine at our scale. With many users, renewing only when the expiry is close would save most writes, but "last used" would become less exact.
- **Kept on purpose:** the conversation list reloads on every live message. It's simple and always right. A very busy group would make it costly; then the list could update one preview from the event instead.
- **Later by design:**
  - the sender's own watermarks (for unread counts);
  - per-device delivered;
  - end-to-end encryption (milestone 8);
  - several copies of `realtime` and `chat` (a good experiment).
- `erasable_syntax=true` in protoc-gen-es is marked "experimental". Watch it when updating.

## Next step

Merged through a pull request once CI passed (the owner had no review notes this time). Try it by hand: `task dev`, two browser windows, alice and bob. Then: the design doc for milestone 4, the first release.
