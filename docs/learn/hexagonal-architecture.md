# Hexagonal architecture (ports and adapters)

## What it is

A way to organize a service's code so the important part, the **business rules**, doesn't depend on any technology. The rules sit in the middle (the **core**). Everything technical sits around it: the database, the API, other services, the message server.

Analogy: a laptop. Inside is the computer (the core). It has **ports**: USB, HDMI, power. It doesn't care whether a mouse or a keyboard is plugged in, as long as it fits the port. The things you plug in are the **adapters**.

```
                 ┌───────────────────────────────┐
  browser ──▶ api (Connect)                       │
                 │   ┌───────────────────────┐   │
                 │   │  core: conversation/  │   │
                 │   │  rules and types      │──▶ Store port ──▶ postgres/ ──▶ Postgres
                 │   │                       │──▶ UserDirectory port ──▶ authclient/ ──▶ auth
                 │   └───────────────────────┘   │
                 └───────────────────────────────┘
```

The name "hexagonal" only comes from how it was first drawn, with six sides. The number doesn't matter.

## Why we use it

`chat` has real rules: only one 1:1 chat per pair, a group has 2 to 50 members, only members see a conversation. Kept apart from SQL and HTTP, they are easy to read and to test, and a technology can be swapped without touching them. For example, NATS for Kafka in milestone 12. See [ADR 0030](../decisions/0030-hexagonal-architecture-for-chat.md).

## Ports and adapters

In `services/chat/`:

| Folder | Role | What it does |
| --- | --- | --- |
| `conversation/` | **core** | Types (`Conversation`, `Member`), rules (`rules.go`), the use cases (`Service`: start a chat, create a group, list) |
| `conversation/ports.go` | **ports** | Interfaces the core needs: `Store` and `UserDirectory` |
| `postgres/` | **adapter** (driven) | Implements `Store` with the chat database and sqlc |
| `authclient/` | **adapter** (driven) | Implements `UserDirectory` by calling auth's internal API |
| `api/` | **adapter** (driving) | Offers the core to browsers as the `chat.v1` Connect API |
| `main.go` | wiring | Creates the adapters and plugs them into the core |

- **Driving** adapters call the core (the API: a request comes in). **Driven** adapters are called by the core (the database: the core needs data).
- **The rule:** imports only point inward. `postgres` and `api` import `conversation`; `conversation` imports none of them. It doesn't even know pgx or Connect exist. Check it yourself: the imports at the top of `conversation/*.go` are only Go's standard library.
- **Errors cross the border too.** The core says "not found" or "invalid request" (`conversation/errors.go`). Only the `api` adapter knows these become the Connect codes `not_found` and `invalid_argument`.

## Tiny example

The core asks for a port, an interface:

```go
// conversation/ports.go
type UserDirectory interface {
	FindUsersByUsername(ctx context.Context, usernames []string) ([]User, error)
}
```

The adapter implements it, here by calling auth:

```go
// authclient/directory.go
func (directory *Directory) FindUsersByUsername(ctx context.Context, usernames []string) ([]conversation.User, error) {
	response, err := directory.client.FindUsersByUsername(ctx, request) // a Connect call to auth
	...
}
```

`main.go` plugs them together:

```go
store := postgres.NewStore(pool)
directory := authclient.NewDirectory(authURL)
conversations := conversation.NewService(store, directory)
```

And a test plugs in fakes instead, with no database and no auth (`conversation/fakes_test.go`).

## Compared with auth's flat style

`auth` keeps everything in one folder: the API handlers, the rules and the SQL. For its size that works, and it has fewer files. But a handler like `LogIn` mixes rate limits, Connect headers, cookies and the database, so testing one rule means setting up all of them. In `chat`, the rule "you can't start a chat with yourself" is tested with two fakes and one call.

The cost: more folders, interfaces and small conversions (a database row → a `conversation.Conversation` → a `chatv1.Conversation`). Worth it when there are real rules. Not for a tiny service like `hello`.

## Alternatives

- **Flat (like `auth`):** fewest files; rules mix with technology as the service grows.
- **Layered (handler → service → repository):** very common. Similar, but the service layer usually imports the repository's code directly, so the rules still depend on the database code.
- **Clean architecture / onion:** the same idea as hexagonal, with more named rings (entities, use cases, interfaces). More ceremony.
- **Domain-driven design (DDD):** about modeling the business with its own words (aggregates, bounded contexts). Often combined with hexagonal; a later topic.

## Learn more

- https://alistair.cockburn.us/hexagonal-architecture/ (the original article, by Alistair Cockburn)
- https://go.dev/doc/modules/layout (Go's own advice on organizing a project)
