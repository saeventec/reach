# Design: Real-time chat

- **Milestone:** m03
- **Status:** Agreed (2026-11-01)
- **Date:** 2026-11-01

## Problem

People who are logged in (milestone 2) can't talk to each other yet. Reach needs chat: a 1:1 conversation between two people, and a group with a name and several members. A message must appear on the other person's screen right away, without a page reload. The sender must see when a message was **delivered** (it reached one of the other person's devices) and when it was **read**.

Reach's [design rules](../PROJECT_BRIEF.md#design-for-the-future) apply from now on: client-made UUIDv7 ids, a sequence number per conversation, an append-only event log, events on NATS, and a server that never needs to read the message text.

## Goals and non-goals

- **Goals:**
  - Start a 1:1 conversation with someone by their username. There is only ever one 1:1 conversation per pair of people.
  - Create a group with a name and its first members.
  - See your conversations, the newest activity first.
  - Send text messages, and see the history (older messages page by page).
  - New messages arrive live, in every open tab and device of each member.
  - Delivered and read status ("✓", "✓✓", "read") for your own messages.
  - Nothing is lost or shown twice: sending the same message twice (a retry) stores it once, and a device that was offline catches up on everything it missed.
- **Non-goals** (each one can be a later step, see [future considerations](../PROJECT_BRIEF.md#future-considerations)):
  - Adding or removing group members after creation, leaving a group, group admins.
  - Editing, deleting, replies, reactions.
  - Typing indicators, online status.
  - Files and pictures (milestone 5), push notifications, end-to-end encryption (milestone 8).
  - Blocking people, or "message requests" from strangers. Anyone with an account can message anyone. Fine for a group of friends; a real risk for a public app.

## Design

Two new Go services, NATS, and a new check in the gateway:

```
                 ┌──────────────────────────────────────────┐
                 │  Browser (React)                         │
                 └──────────────────────────────────────────┘
   Connect (HTTP) │ send, list, mark read      ▲ WebSocket: live events
                  ▼                            │
                 ┌──────────────────────────────────────────┐
                 │  Gateway (Caddy)                         │
                 │  1. asks auth: "whose cookie is this?"   │──── forward_auth ───┐
                 │  2. adds X-User-Id, then forwards        │                     ▼
                 └──────────────────────────────────────────┘               ┌────────┐
                   │ /chat.v1.*                      │ /realtime           │  auth  │
                   ▼                                 │                     └────────┘
             ┌───────────┐  usernames → ids (RPC)    │                         ▲
             │   chat    │───────────────────────────┼─────────────────────────┘
             └───────────┘                           ▼
               │      │ outbox relay           ┌────────────┐
               ▼      │ publishes events       │  realtime  │  keeps the open
        ┌──────────┐  ▼                        └────────────┘  WebSockets, per user
        │ Postgres │ ┌──────────────────────┐        ▲
        │  (chat)  │ │  NATS (JetStream)    │────────┘ subscribes to chat.>
        └──────────┘ └──────────────────────┘
```

### 1. Who is asking: authentication at the gateway

Today only `auth` reads the session cookie. `chat` and `realtime` must know the user too. **Proposal:** Caddy's `forward_auth`. Before Caddy forwards a request to `chat` or `realtime`, it sends a copy of the request (with the cookie) to a small endpoint in `auth`. If the session is valid, `auth` answers `200` with an `X-User-Id` header, and Caddy adds it to the real request. If not, the caller gets `401` and the request never reaches `chat`.

- `chat` and `realtime` stay simple: they read one header and never see the cookie.
- Big companies do the same in their API gateways (Envoy's "external authorization").
- **Security:** a service must only trust `X-User-Id` from the gateway. So (1) Caddy removes any `X-User-Id` the browser sent, before anything else, (2) `chat` and `realtime` have no published ports, so only the gateway reaches them, and (3) an integration test sends a fake `X-User-Id` through the gateway and expects it to be refused. Caddy's docs don't say whether `forward_auth` removes a forged header by itself, so we remove it ourselves.
- Cost: one extra call to `auth` (and one database read) per request. Fine at our scale; a cache is a later topic.

### 2. Where the data lives: the `chat` service and its database

`chat` owns conversations and messages, in its own `chat` database (ADR 0024). Four tables:

| Table | One row per | Main columns |
| --- | --- | --- |
| `conversations` | conversation | `id`, `kind` (direct or group), `name` (groups only), `direct_pair_key` (direct only, unique), `last_sequence` |
| `members` | person in a conversation | `conversation_id`, `user_id`, `delivered_up_to`, `read_up_to` |
| `messages` | message (append-only) | `id` (made by the client), `conversation_id`, `sequence`, `sender_id`, `text`, `sent_at` |
| `outbox` | event waiting to be published | `id`, `subject`, `payload` (Protobuf bytes), `created_at`, `published_at` |

- **Usernames:** `chat` stores user ids only. To turn "bob" into an id (and an id back into a name for the screen), it calls `auth` with Connect RPC: a new internal `authinternal.v1.UserDirectoryService`. Caddy routes only `/auth.v1.*`, so the browser can't reach it. (Changed while building: the first idea, `auth.internal.v1`, would put the Go code in a folder named `internal`, which Go lets only nearby code import.)
- **One 1:1 per pair:** `direct_pair_key` is the two user ids, sorted and joined. A `UNIQUE` constraint means a second "start a chat with bob" finds the existing conversation.

### 3. Order: a sequence number per conversation

Each message gets the next number in its conversation: 1, 2, 3… This is the order everyone shows, and a device catches up by asking "everything after 41".

- Sending runs in **one transaction**: `UPDATE conversations SET last_sequence = last_sequence + 1 … RETURNING last_sequence`, then insert the message with that number, then insert the event into `outbox`. The `UPDATE` locks the conversation's row, so two messages sent at the same moment get 42 and 43, never two 42s.
- Why not order by time: two computers' clocks never agree exactly, and two messages can have the same time.

### 4. Retries: client-made ids (idempotency)

The browser creates each message's id (UUIDv7) before sending. If the network drops and the browser sends again, the id is the same, so the server finds the existing message and returns it instead of storing a copy. **Idempotent** means: doing it twice has the same effect as doing it once.

### 5. Events: NATS JetStream and the transactional outbox

When something happens, `chat` publishes an event: `chat.message.sent`, `chat.conversation.created`, `chat.receipt.updated`. Events are Protobuf messages in `proto/chat/v1/events.proto`. Each event carries the ids of the conversation's members, so `realtime` knows who to deliver it to without asking `chat`.

The problem: saving to Postgres and publishing to NATS are two different systems. If `chat` saves the message and then crashes before publishing, the message exists but nobody is told. **The transactional outbox** fixes this: the event is saved in the `outbox` table *in the same transaction* as the message. A small loop in `chat` (the "relay") reads unpublished rows, publishes them, and marks them published. Either both the message and its event are saved, or neither.

- The relay can publish an event twice (published, then crashed before marking it). So events are delivered "at least once", and receivers must ignore duplicates: the browser ignores a sequence number it already has.
- The relay can be woken right away after a send (a Go channel), and also checks every second, so events are fast and never stuck.
- The event code sits behind a small Go interface (`outbox.Publisher`; first called `eventPublisher` here), so Kafka can replace NATS later (milestone 12).
- Added while building (step 8): each event is published with its outbox id as the JetStream message id (`Nats-Msg-Id`), so JetStream itself drops most copies. The stream `CHAT` keeps `chat.>` for 24 hours, because the history is in Postgres. The relay reads rows without locking them (no `FOR UPDATE SKIP LOCKED`): with one copy of `chat` there is nothing to share, and with two, the message id drops the copies.
- Each event comes with the step that first needs it: `chat.message.sent` in step 8, `chat.receipt.updated` in step 10, `chat.conversation.created` in step 11 (a new chat appears live in the list).

### 6. Live delivery: the `realtime` service and WebSocket

- The browser opens one WebSocket to `/realtime` after login. Caddy checks the cookie first (section 1), so `realtime` knows the user.
- `realtime` keeps a map in memory: user id → that user's open connections (each tab is one). This is not important data: after a restart, browsers reconnect.
- It subscribes to `chat.>` on NATS. For each event, it sends a copy to every open connection of every member.
- `realtime` has no database. Many goroutines touch the map at once, so a mutex protects it, and tests run with Go's race detector (`go test -race`).
- **Library:** [coder/websocket](https://github.com/coder/websocket): maintained, small, `context` support, safe for writes from several goroutines, and it refuses requests from other websites by default (it checks the `Origin` header). That stops "cross-site WebSocket hijacking", where another website opens a socket with your cookie.
- Added while building (step 9): each WebSocket message is a `realtime.v1.ServerEvent` as JSON (a `oneof`, so new kinds of updates become new fields). The hub never waits for a slow browser: each WebSocket has a buffer of 32 events, and a browser whose buffer is full is closed with "try again later", reconnects and catches up. On a stop, each WebSocket is closed with "going away". The race detector needs a C compiler, so it runs in its own `task test:race` (and in CI), not in `task test`.
- **Catching up:** the WebSocket is only for speed. When a browser connects (or reconnects after being offline), it calls `ListMessages(after: N)` for the conversation it shows. If it receives message 45 but has only up to 42, it fetches 43 and 44. So a lost event is never a lost message.

### 7. Delivered and read

Each member row has two numbers ("watermarks"): `delivered_up_to` and `read_up_to`. "Read up to 42" means messages 1 to 42 are read. One number per member is much smaller than one row per message per person, and it fits the sequence numbers.

- **Delivered:** when a browser receives messages (live or by catching up), it calls `MarkDelivered(conversation, 42)`.
- **Read:** when the conversation is open on screen and visible, it calls `MarkRead(conversation, 42)`.
- The numbers only go up (`GREATEST(old, new)` in SQL), so an old, late call can't move them back.
- Each change is an event (`chat.receipt.updated`), so the sender's screen updates live.
- Per person, not per device: "delivered" means any of bob's devices got it. Enough for now; the data model still lets us add per-device later.
- In a group, "✓✓" means delivered to everyone, and "read" means read by everyone.

### 8. The `chat` API (Connect, `chat.v1`)

| Call | What it does |
| --- | --- |
| `StartDirectConversation(username)` | Finds or creates the 1:1 with that person |
| `CreateGroup(name, usernames)` | Creates a group with you and them |
| `ListConversations()` | Your conversations, newest activity first, with the last message |
| `SendMessage(conversation_id, message_id, text)` | Stores a message (idempotent) |
| `ListMessages(conversation_id, after or before, limit)` | History: newer than N (catching up) or older than N (scrolling up) |
| `MarkDelivered(conversation_id, sequence)` / `MarkRead(…)` | Moves your watermarks |

Every call checks that you are a member of the conversation. If you are not, the answer is `NOT_FOUND`, never "it exists but it isn't yours".

### 9. Shared Go code

Milestone 2 decided: shared code moves to `packages/go/` when a third Go service arrives. `chat` is the third (and `realtime` the fourth). So the first step moves the common parts (reading settings, the health check, connecting to Postgres, starting and stopping the server cleanly) into one small Go module, `packages/go`. Each service points to it with a `replace` line in its `go.mod`. Docker then needs the whole repo as its build context, not only the service's folder.

### 10. Code structure of `chat`

`auth` is "flat": all files in one folder. That's fine for a small service, but in `chat` the business rules ("only members may send", "watermarks only go up") would get mixed with SQL, Connect and NATS code. **Proposal: hexagonal architecture** ("ports and adapters"), kept small:

```
services/chat/
├── main.go            # wires everything together
├── conversation/      # the core: rules and types, no SQL, no HTTP, no NATS
│   ├── service.go     #   "send a message", "mark read", ...
│   └── ports.go       #   interfaces the core needs: a store, a user directory
├── postgres/          # adapter: the store, with sqlc
├── api/               # adapter: the Connect handlers
├── outbox/            # adapter: the relay that publishes to NATS
└── authclient/        # adapter: the user directory, calling auth
```

The core only knows its interfaces ("ports"). So its tests need no database, and a part (NATS → Kafka) can be swapped without touching the rules (Golden rule 16). We compare it with `auth`'s flat style in a learning note.

### 11. The web app

- `/chats`: your conversations, a "new chat" and a "new group" form.
- `/chats/:id`: the messages, a text box, and the status of your own messages.
- One WebSocket for the whole app, opened after login and reopened after a drop (with a growing wait between tries: "backoff").
- A message appears right away as "sending…" (optimistic update), and gets its sequence number when the server confirms.
- UUIDv7 in the browser: `crypto.randomUUID()` makes only version 4, so we add the small, widely used [`uuid`](https://www.npmjs.com/package/uuid) library (or write the 20 lines ourselves, to learn the format).

### 12. Tests

- Unit tests for the core rules (no database).
- Integration tests with Testcontainers: Postgres (as in milestone 2) and now NATS: send a message, and see the event arrive.
- `realtime` tests with the race detector.
- End to end (proposal): **Playwright** opens two browsers, logs in as alice and bob, and checks that bob sees alice's message live, and that alice then sees "read".

## Options considered

**Authentication for `chat` and `realtime`:**

- **A. Caddy `forward_auth`, `auth` answers with the user id (proposed):** one place checks sessions, services stay simple. Con: every service trusts a header, so the network must be closed.
- **B. Each service calls `auth` itself:** no trusted header, but every service repeats the same code.
- **C. Signed tokens (JWT) checked by each service:** no call to `auth` per request, but a logged-out device stays valid until its token expires, and keys must be managed. A good later topic.

**Live delivery:**

- **A. WebSocket (proposed, in the brief):** two-way, in every browser, the industry standard for chat.
- **B. Connect server streaming:** stays inside our Protobuf API with no new library, but one direction only and less common for chat.
- **C. Server-Sent Events:** simple, one direction only, plain HTTP.
- **D. Polling every few seconds:** simplest, but slow and wasteful.

**Publishing events:**

- **A. Transactional outbox (proposed):** never "saved but not announced". Con: a table and a relay loop.
- **B. Publish right after the commit:** simplest, but a crash in between loses the event (the browser still catches up on its next fetch, so it would be "late", not "lost").
- **C. Read Postgres's change log (CDC, for example Debezium):** no outbox code, but a big extra system.

**Ordering:**

- **A. A counter on the conversation row (proposed):** simple, exact, one row lock per conversation.
- **B. Timestamps:** clocks disagree, duplicates happen.
- **C. One global Postgres sequence:** no lock, but gaps and no "per conversation" order.

**Code structure of `chat`:**

- **A. Hexagonal (proposed):** clear core, easy tests and swaps. Con: more folders and interfaces.
- **B. Flat, like `auth`:** less code, but rules mix with SQL and transport.
- **C. Layered (handler → service → repository):** common and simple, but the core still depends on the database code.

**Sharing Go code:** `replace` lines in each `go.mod` (proposed: works the same in Docker and locally), a `go.work` file (Go's workspace feature: nicer locally, but awkward in each Docker build), or keep copying (two copies are fine, four drift apart).

## Trade-offs and risks

- **Bigger than milestone 2.** Two services, NATS and a new web UI. The steps are kept small, each one runs.
- **Trusted header.** If a service ever gets a published port, anyone could fake `X-User-Id`. The Compose comments and a test guard this.
- **One row lock per conversation.** A very busy group would wait on that lock. Not a problem for friends; a good topic to measure later (load testing).
- **Events at least once.** Every receiver must handle duplicates. The browser handles them by sequence number.
- **Several copies of `realtime`.** It works with 2 copies, because each copy receives every event and delivers only to its own sockets. A good small experiment after this milestone.
- **Message text in plain text in Postgres.** Until end-to-end encryption (milestone 8). The design never needs to read it, so it can become encrypted later.
- **No abuse protection yet**, apart from a size limit (4000 characters per message; changed while building from "4 KiB", because people count characters, not bytes) and the membership check. Rate limiting message sending is a small later step.

## Learning

From the [learning map](../PROJECT_BRIEF.md#learning-map):

- **Networking:** WebSocket, the HTTP upgrade, keep-alive pings, reconnecting with backoff.
- **Distributed systems:** idempotency, ordering, at-least-once delivery, the transactional outbox, eventual consistency.
- **Databases, in depth:** transactions, row locks, `UPDATE … RETURNING`, constraints for business rules.
- **Concurrency:** goroutines, mutexes, the race detector (in `realtime`).
- **Clean architecture:** hexagonal architecture in `chat`, compared with `auth`.
- **System design:** an API gateway that authenticates, fan-out to many connections.
- **Security:** trusted headers, cross-site WebSocket hijacking, authorization ("is this user a member?").
- **Testing strategy:** end-to-end tests in a real browser.
- **Frontend engineering:** optimistic updates, live state.

## Decisions

The owner chose options A for authentication (ADR 0029), code structure (ADR 0030) and events (ADR 0031), and said yes to Playwright (ADR 0034). The other choices are the agent's proposals: ADRs 0032, 0033 and 0035.

## Open questions

- UUIDv7 in the browser: the `uuid` library or our own small function. Decided in step 12.
- Running 2 copies of `realtime` to see the fan-out: a small experiment after this milestone, if the owner wants it.
