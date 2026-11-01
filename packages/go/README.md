# packages/go

Small Go code shared by every Go service: reading settings, the health check, connecting to Postgres, and starting and stopping the HTTP server cleanly. See [ADR 0035](../../docs/decisions/0035-share-go-code-in-packages-go.md).

| Package | What it does |
| --- | --- |
| `config` | Reads settings from environment variables |
| `health` | The `/healthz` handler |
| `database` | Connects to Postgres and checks that it answers |
| `httpserver` | Runs an HTTP server until a stop signal, then shuts it down cleanly |
| `gen` | Go code generated from `proto/` by `task generate`. Never edit it by hand |

Keep it small: only code that several services really share, with tests. A change here affects every service.

Learn more: [docs/learn/go.md](../../docs/learn/go.md#sharing-code-between-modules)
