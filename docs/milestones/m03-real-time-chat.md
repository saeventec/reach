# Milestone 3: Real-time chat

- **Status:** In progress (started 2026-11-01), on the branch `milestone-3-chat`. Design: [m03-real-time-chat.md](../design/m03-real-time-chat.md).

## Goal

Logged-in people can start a 1:1 chat or a group, send text messages, and see new messages arrive live. Senders see when a message was delivered and read. Two new services do this: `chat` (conversations and messages, with its own database) and `realtime` (WebSocket connections), connected by NATS.

```
Browser → Caddy (checks the session with auth) → chat → Postgres (chat)
                                                   │ outbox → NATS
Browser ⇄ WebSocket ⇄ Caddy ⇄ realtime ◀───────────────────┘
```

## Steps

Each step is one small commit that runs. The agent commits each finished step right away, and the owner reviews the commits afterwards.

- [x] 1. Plan: the design doc, this file and the decisions below
- [x] 2. Shared Go code: move config, health check, database connection and clean shutdown into `packages/go`; `hello` and `auth` use it (ADR 0035)
- [x] 3. NATS with JetStream in Compose, with a health check
- [x] 4. Authentication at the gateway: a session-check endpoint in `auth`, and `forward_auth` in Caddy that removes forged `X-User-Id` headers (ADR 0029)
- [x] 5. The `chat` service skeleton, in hexagonal style: its database, user and Compose entries (ADR 0030). For now only `main.go` and the health check: the hexagonal folders come with their first code in step 6, so no folder is empty.
- [x] 6. Conversations: the first migrations, `StartDirectConversation`, `CreateGroup` and `ListConversations`; usernames from a new internal `auth` API. Built in three commits: all generated Go code moved to `packages/go/gen` first (ADR 0017), then the internal API, then the conversations.
- [x] 7. Messages: `SendMessage` with sequence numbers and client-made ids, and `ListMessages` (ADR 0032). The last message in `ListConversations` comes with the conversation list page (step 11), where it is first needed.
- [x] 8. Events: the outbox table and the relay that publishes to NATS JetStream (ADR 0031). Only `chat.message.sent` for now: `chat.receipt.updated` comes with step 10, and `chat.conversation.created` with step 11, where each is first needed.
- [x] 9. The `realtime` service: WebSocket connections, subscribing to NATS, pushing events to each member's open connections (ADR 0033). Built in two commits: the NATS address and the event subjects moved to `packages/go` first, then the service. The race detector runs in its own `task test:race` and in CI.
- [x] 10. Delivered and read: `MarkDelivered`, `MarkRead` and their event (`chat.receipt.updated`). Built in two commits: chat first, then realtime forwarding the event.
- [x] 11. The web app: the conversation list, new chat and new group forms; the last message in `ListConversations`, and the `chat.conversation.created` event. Built in five commits: the last message, the event, a fix for generated TypeScript enums (`task build` had failed since step 6), the list page, and live updates. The list items link nowhere yet: the conversation page comes in step 12.
- [ ] 12. The web app: the conversation page, live updates, catching up, delivered and read status
- [ ] 13. End-to-end tests with Playwright: alice and bob chat live, locally and in CI (ADR 0034)
- [ ] 14. Security and tests: a message size limit, the forged-header test, integration tests with Postgres and NATS, the race detector on `realtime`

The steps can change as we learn. Any change is written here.

## Decisions

- [0029](../decisions/0029-authenticate-at-the-gateway.md): authenticate at the gateway with `forward_auth` (the owner's choice)
- [0030](../decisions/0030-hexagonal-architecture-for-chat.md): hexagonal architecture for `chat` (the owner's choice)
- [0031](../decisions/0031-transactional-outbox-for-chat-events.md): a transactional outbox for chat events (the owner's choice)
- [0032](../decisions/0032-order-messages-by-sequence-per-conversation.md): a sequence number per conversation, client-made ids, read and delivered as watermarks
- [0033](../decisions/0033-use-coder-websocket.md): coder/websocket in `realtime`
- [0034](../decisions/0034-use-playwright-for-end-to-end-tests.md): Playwright for end-to-end tests (the owner's choice)
- [0035](../decisions/0035-share-go-code-in-packages-go.md): shared Go code in `packages/go`, with `replace` lines
- Not in this milestone: changing group members after creation, edits, deletes, reactions, typing indicators, online status, blocking (see the design doc's non-goals).

## Concepts learned

- Filled in as the steps are built, with links to `docs/learn/`.

## Done when

`task dev` starts everything. In two browsers, alice and bob can start a chat, send messages that arrive live, and see "delivered" and "read". A group with three people works the same way. After a reload or a short time offline, nothing is missing and nothing is shown twice. Tests, the Playwright tests and CI pass.
