# WebSocket and the realtime service

## What it is

Normal HTTP is "ask, then get an answer": the browser asks, the server answers, done. The server can never start talking on its own.

A **WebSocket** is a connection that stays open, and both sides can send messages on it at any time. Analogy: HTTP is sending letters; a WebSocket is a phone call that stays connected.

A WebSocket starts as a normal HTTP request with two special headers: `Connection: Upgrade` and `Upgrade: websocket`. The server answers `101 Switching Protocols`, and from then on, the same connection carries WebSocket messages instead of HTTP. That's the **upgrade**, or the "handshake". Because it starts as HTTP, the browser sends its cookie, and the gateway can check the session as for any request.

In the browser:

```ts
const socket = new WebSocket("/realtime"); // the browser makes it ws://localhost:8080/realtime
socket.onmessage = (message) => console.log(message.data);
```

## Why we use it

When alice sends a message, bob's screen should show it at once, without bob asking. Only the server knows when something happens, so the server must be able to speak first. See the [milestone 3 design](../design/m03-real-time-chat.md#6-live-delivery-the-realtime-service-and-websocket).

## How the realtime service works

```
alice ── SendMessage ──▶ chat ── outbox ──▶ NATS (chat.message.sent)
                                               │
                                               ▼ subscribed to chat.>
bob's tab 1 ◀── WebSocket ──┐            ┌──────────────┐
bob's tab 2 ◀── WebSocket ──┼── Caddy ◀──│  realtime    │  the hub: bob → [tab 1, tab 2]
carol       ◀── WebSocket ──┘            │  (no database)│           carol → [tab 1]
                                         └──────────────┘
```

The code is in [services/realtime/](../../services/realtime/). It is a small service with no business rules, so its files sit in one folder, like `auth` (compare with `chat`'s [hexagonal style](hexagonal-architecture.md)).

| File | What it does |
| --- | --- |
| `main.go` | Connects to NATS, subscribes, and runs the HTTP server |
| `socket.go` | `GET /realtime`: accepts a WebSocket and writes the user's events to it |
| `hub.go` | The hub: who is connected right now |
| `events.go` | Turns a NATS event into a WebSocket message, and gives it to the hub |

### From NATS to the browser

1. `realtime` subscribes to `chat.>` with a plain ("core") NATS subscription. Each copy of `realtime` gets every event.
2. For `chat.message.sent`, `events.go` reads the Protobuf event. It contains the conversation's member ids (see [the transactional outbox](transactional-outbox.md)), so `realtime` never asks `chat` who is in a conversation.
3. It wraps the event (`MessageSent`, `ReceiptUpdated` or `ConversationCreated`) in a `realtime.v1.ServerEvent` ([proto/realtime/v1/events.proto](../../proto/realtime/v1/events.proto)) and turns it into JSON, the same JSON the Connect APIs use. Each WebSocket message is one `ServerEvent`: `{"messageSent": {"message": {...}, "memberIds": [...]}}`. A `oneof` says which kind of update it is, so new kinds become new fields.
4. The hub hands the JSON to every open WebSocket of every member.

Events published while `realtime` was down are simply missed. That's fine: the WebSocket is only for speed. The browser catches up with `ListMessages(after: N)` when it (re)connects, so a lost event is never a lost message.

### The hub

The hub is a map from a user id to that user's open WebSockets: one per browser tab. Each WebSocket has a small **buffer**, a channel with room for 32 events.

- **Delivering never waits.** The NATS subscription calls the hub for every event, one at a time. If it waited for a slow browser, everyone would wait. So the hub puts the event in each buffer without waiting (`select` with `default`). If a buffer is full, that browser is far behind: the hub drops it, and its WebSocket is closed with "try again later". The browser reconnects and catches up.
- **Each WebSocket has its own loop** (`sendEvents` in `socket.go`): it takes events from its buffer and writes them, with a 10-second limit per write.

### Mutexes

Many goroutines use the hub at the same time: one per WebSocket (adding and removing itself), plus the NATS subscription (delivering). A Go map must never be changed by one goroutine while another one uses it. That's a **data race**, and it can corrupt the map or crash the program.

A **mutex** (`sync.Mutex`, "mutual exclusion") is a lock: `Lock()` waits until no one else holds it, and `Unlock()` lets the next one in. Every hub method locks it first:

```go
func (hub *Hub) add(newClient *client) {
	hub.mutex.Lock()
	defer hub.mutex.Unlock() // unlock however the function returns
	...
}
```

The work inside the lock is tiny (a map lookup, a channel send that never waits), so nobody waits long.

### The race detector

A data race is hard to see: the program works most of the time, and fails rarely and randomly. Go's **race detector** watches every memory access while the tests run, and reports two goroutines touching the same memory without a lock between them.

```
task test:race     # runs: go test -race ./...  in services/realtime and services/chat
```

`TestHubIsSafeForManyGoroutines` in [hub_test.go](../../services/realtime/hub_test.go) adds, removes and delivers from 100 goroutines at once. We checked it: with the lock taken out of `Deliver`, the race detector reports 9 data races and the test fails.

The race detector needs a C compiler. A Mac has one (Xcode's command line tools); on a new WSL2 computer, install one with `sudo apt install gcc`. So it is a separate task, not part of `task test`. CI runs it on every push.

### The origin check

The browser sends the cookie with every request to our site, also a WebSocket opened by **another website's** page. Without a check, `evil.example` could open `ws://localhost:8080/realtime` in your browser and read your messages: "cross-site WebSocket hijacking".

The browser always says which page opened the WebSocket, in the `Origin` header, and a page can't change it. `websocket.Accept` from [coder/websocket](https://github.com/coder/websocket) refuses an `Origin` that doesn't match the address it connects to, with `403 Forbidden`. That is its default; we set no options. A test checks it.

### Stopping cleanly

`http.Server.Shutdown` waits for open requests, but not for WebSockets: once a request has become a WebSocket, the server no longer tracks it ("hijacked"). So `realtime` does it itself:

1. On the stop signal, each WebSocket's loop closes it with **"going away"** (status 1001). The browser knows to reconnect.
2. A `sync.WaitGroup` counts the open WebSockets. `main.go` waits for it to reach zero after the HTTP server has shut down.

We found one trap while testing. `CloseRead` (which reads in the background, to notice when the browser leaves) takes a context. When that context ends, the library cuts the connection at once, with no goodbye. So the stop signal must not be its context; the loop watches the stop signal separately.

## In the browser

[apps/web/src/chat/liveUpdates.ts](../../apps/web/src/chat/liveUpdates.ts) keeps the WebSocket open for the web app:

- **The address:** `ws://localhost:8080/realtime`, the page's own address with `ws:` instead of `http:` (`wss:` with HTTPS). The browser sends the cookie with it.
- **Reading:** each message is JSON, read into a `ServerEvent` with `fromJsonString`. `ignoreUnknownFields` makes it skip fields a newer server adds, instead of failing.
- **Reconnecting:** when the WebSocket closes (the server restarted, the network dropped), it tries again after 1 second, then 2, 4, 8 ... up to 30 seconds. That's **exponential backoff**: when the server is down, a thousand browsers don't all knock every second. A successful connect starts over at 1 second.
- **Catching up:** every time it connects, also the first time, it calls `onConnected`. The page then loads what it shows again, because events sent while the WebSocket was closed are lost.
- **The conversation list** ([ConversationListPage.tsx](../../apps/web/src/chat/ConversationListPage.tsx)) loads the list again on `messageSent`, `conversationCreated` and every connect. Simple and always right: the server sorts the list and writes the previews. Its effect's cleanup closes the WebSocket when the page closes.
- **Tests** use a fake socket and **fake timers** (`vi.useFakeTimers()`): the test moves the clock itself, so "wait 30 seconds" takes no time.

For now, each page that wants live updates opens its own WebSocket. Sharing one between pages is a small later improvement.

## What we tested

- **Unit tests** with a real local server and a real WebSocket client (`httptest.NewServer` and `websocket.Dial`): events arrive; no `X-User-Id` gives 401; another website's page gives 403; a closed tab leaves the hub; stopping says "going away"; a browser that falls behind is closed.
- **NATS integration test**: an event published on `chat.message.sent` reaches the right user's buffer.
- **By hand, with the whole system through the gateway:** both of bob's tabs got alice's message, as JSON, and carol got nothing.
  - No cookie: 401. Another website: 403.
  - bob, logged in, sending `X-User-Id` with carol's id: the gateway replaced it, and bob got none of carol's messages.
  - Restarting `realtime`: bob got "going away", and the service stopped in half a second.

## Tiny example

The heart of `socket.go`:

```go
connection, err := websocket.Accept(writer, request, nil) // the upgrade, with the origin check
...
for {
	select {
	case <-closed.Done():           // the browser left
		return
	case frame := <-userClient.frames: // an event for this user
		connection.Write(ctx, websocket.MessageText, frame)
	}
}
```

## Alternatives

- **Server-Sent Events (SSE):** the server streams messages over normal HTTP, one direction only. Simpler, and enough for us today, since the browser sends nothing on the socket. WebSocket leaves room for typing indicators later.
- **Connect server streaming:** stays inside our Protobuf APIs, but one direction only, and less common for chat.
- **Polling:** the browser asks "anything new?" every few seconds. Simplest, but slow and wasteful.
- **Libraries:** gorilla/websocket (famous, no `context` support), or writing it ourselves on `net/http` (a good exercise, much protocol detail). See [ADR 0033](../decisions/0033-use-coder-websocket.md).

## Learn more

- https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API
- https://pkg.go.dev/github.com/coder/websocket
- https://go.dev/doc/articles/race_detector
- https://go.dev/tour/concurrency/9 (sync.Mutex)
