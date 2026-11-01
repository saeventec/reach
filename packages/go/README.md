# packages/go

Small Go code shared by the Go services: reading settings, the health check, connecting to Postgres, starting and stopping the HTTP server cleanly, and the names of chat's events. See [ADR 0035](../../docs/decisions/0035-share-go-code-in-packages-go.md).

| Package | What it does |
| --- | --- |
| `config` | Reads settings from environment variables |
| `health` | The `/healthz` handler |
| `database` | Connects to Postgres and checks that it answers |
| `httpserver` | Runs an HTTP server until a stop signal, then shuts it down cleanly |
| `chatevents` | The NATS subjects of chat's events, for chat (publishes) and realtime (subscribes) |
| `gen` | Go code generated from `proto/` by `task generate`. Never edit it by hand |

Keep it small: only code that several services really share, with tests. A change here affects every service.

Learn more: [docs/learn/go.md](../../docs/learn/go.md#sharing-code-between-modules)
