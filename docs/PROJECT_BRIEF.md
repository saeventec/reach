# Reach — Project Brief & Agent Guide

Oct 30, 2026 · @Sadra Samadi

## How to use this document

The reference for what Reach is and where it is going: goals, scope, architecture, stack and milestones.

- **Read it by section, when a task needs it.** It is not read at the start of every session ([AGENTS.md](../AGENTS.md#reading-docs)).
- **The rules** (golden rules, code rules, workflow, where the docs are) live only in [AGENTS.md](../AGENTS.md).
- **If something here is unclear or seems wrong:** stop and ask the owner. Never guess.
- **If a decision changes:** update this document and write an ADR in `docs/decisions/`.

## The goal

Reach is a communication app built from scratch, so the owner learns what a senior software engineer knows, and shows it (ADR 0028).

- **What Reach does:** chat (1:1 and groups), voice calls, video calls, and file sharing.
- **Online and offline:** online through our own server. Later, in the native apps, also offline over local networks (Wi-Fi, LAN, Bluetooth), with an optional sync to the server when back online.
- **Platforms:** web first. Later Windows, macOS, Linux, Android, iOS and iPadOS.
- **Who uses it:** for now, the owner and their friends. Not a commercial product, but real people use it, so security and privacy are taken seriously.
- **The real goal is learning.** The app is the vehicle. The owner wants to learn the concepts, tools, standards, architectures and practices that engineers use for enterprise-grade software in large companies: infrastructure, DevOps, backend, frontend, APIs, databases, security, networking, operating systems, messaging (NATS, later Kafka), containers (Docker, later Kubernetes), CI/CD, observability, clean code, clean architecture, versioning and release strategy. More areas are in the [learning map](#learning-map).
- **A showcase.** The repo is the owner's portfolio: others read it to see what the owner can do. So the code, the docs and every decision (with its reasons and trade-offs) must be easy to read and defend.
- **Real concepts, tiny scale.** Advanced topics are welcome, but small and simple, like a working scale model. Examples: 2 or 3 copies of a service instead of hundreds of servers; Kubernetes with k3s on the owner's computer instead of a cloud; a tiny version of a hard system (a sync protocol, a media server) to understand why the big tools exist. A topic is never skipped only because Reach is small: it becomes a planned, clearly labeled **learning module**, and its ADR says honestly what it teaches and what it would cost in a real project.
- **Observe first, direct later.** For now the owner doesn't write code: the agent builds and explains everything, and the owner reviews and learns each step before moving on. Later, with the full picture, the owner directs the structure, design and style.
- **Understanding beats speed.** The owner must understand every file and every line in the project. A slower, clearer solution always wins over a faster, clever one.
- **The repo is public** on GitHub.

## Product scope and phases

We start with an online-only web app, and add native apps and offline features later.

| Phase | Clients | Connectivity | Features |
| --- | --- | --- | --- |
| 1. Web | Web app (React) | Online only, always through the server | Accounts, chat (1:1, groups), files, 1:1 voice and video calls |
| 2. Native | Desktop (Windows, macOS, Linux) and mobile (iOS, iPadOS, Android) | Online, plus a local copy of the user's data | Everything from phase 1, plus local storage and sync |
| 3. Offline | Native apps only | Local Wi-Fi and LAN, then Bluetooth | Chat, files and calls between nearby devices with no internet |
| 4. Scale | All | All | Group calls, end-to-end encryption, multi-device, releases |

**Offline sync (owner's idea).** Native apps keep a copy of the user's data in a local database. Messages created offline (also those exchanged directly between devices over LAN or Bluetooth) are uploaded to the server when the device is back online, but only if the user wants that. Each conversation can be set to "sync to server" or "this device only".

**Why the web is online-only:** browsers are not allowed to discover devices on a local network or open Bluetooth connections to other devices, so offline features only work in native apps.

**Push notifications:** allowed through Apple (APNs) and Google (FCM), since iOS has no self-hosted alternative.

## Architecture

The browser talks to one gateway, which routes each request to a small Go service; services keep their data in Postgres and share events through NATS. Everything runs as containers in Docker Compose.

```
Reach architecture · phase 1 (web), with later parts

            ┌──────────────────────────────┐
            │  Browser: web app (React)    │
            └──────────────────────────────┘
               │ HTTP (Connect RPC)   ▲ WebSocket
               ▼                      │
            ┌──────────────────────────────┐
            │  Gateway (Caddy)             │
            └──────────────────────────────┘
               │            │            │
               ▼            ▼            ▼
          ┌────────┐   ┌────────┐   ┌──────────┐
          │  auth  │   │  chat  │   │ realtime │   Go services
          └────────┘   └────────┘   └──────────┘
               │          │    │          ▲
               ▼          ▼    │ publish  │ subscribe
          ┌─────────────────┐  ▼          │
          │   PostgreSQL    │ ┌───────────────────┐
          │ (one database   │ │  NATS (JetStream) │
          │  per service)   │ └───────────────────┘
          └─────────────────┘

  All boxes above run as containers in Docker Compose.

  Later milestones add:
  ┌──────────────┐  ┌────────────────────┐  ┌──────────────────────┐
  │ File storage │  │ Calls (WebRTC,     │  │ Observability        │
  │ (S3-style)   │  │ coturn)            │  │ (metrics, logs,      │
  └──────────────┘  └────────────────────┘  │ traces, dashboards)  │
                                            └──────────────────────┘
```

The web app only knows the gateway's address. The chat service stores a message and publishes an event; the realtime service receives it and pushes it to the right browsers over WebSocket. The parts at the bottom are added in later milestones.

## Tech stack decisions

These are the current choices. When an agent introduces any of them for the first time, it explains: what it is, why we chose it, the alternatives, and a tiny example ([golden rules 2 and 4](../AGENTS.md#golden-rules)).

**Used from the start**

| Area | Choice | What it is, in one line | Why we chose it | Alternatives |
| --- | --- | --- | --- | --- |
| Version control | Git + GitHub | Records every change; GitHub hosts the repo and runs CI | Industry standard; owner's choice | GitLab, Forgejo |
| Containers | Docker + Docker Compose | Packs each program into an isolated box; Compose starts them all together | Whole system runs on one computer with one command; same containers go to Kubernetes later | Podman; OrbStack (Mac runtime) |
| Backend language | Go | Small, fast language made for servers | Easy to learn fully, great at many connections, compiles to one small binary | Node.js/TypeScript, Rust, Kotlin, C# |
| Frontend language | TypeScript | JavaScript with types | Catches mistakes early; standard for serious web apps | JavaScript, Dart |
| Web UI | React + Vite | React builds pages from components; Vite runs and bundles the app | Most used UI library, huge learning material; Vite needs almost no config | Svelte, Vue, Solid |
| API contract | Protobuf + Connect RPC | Define messages and API once, generate Go and TypeScript code | Front and back can never disagree; also speaks JSON for easy debugging; same format reused for offline transports later | REST + OpenAPI, gRPC, GraphQL |
| Real-time | WebSocket | An open connection so the server can push messages instantly | Supported in every browser; simple | Server-Sent Events, polling, WebTransport |
| Server database | PostgreSQL | Table-based database queried with SQL | Reliable, free, powerful, very popular | MySQL/MariaDB, MongoDB, CockroachDB |
| DB migrations | goose | Versioned SQL files that change the database step by step | Plain SQL you can read | golang-migrate, Atlas |
| Service messaging | NATS (with JetStream) | A "post office": services publish and subscribe to topics | Tiny, simple, fast; JetStream can keep messages | Kafka (later), RabbitMQ, Redis Pub/Sub |
| Gateway | Caddy | One entry point that routes browser requests to the right service | Tiny config, automatic HTTPS later on the server | Traefik, Nginx, Envoy |
| Tool versions | mise | Installs the exact tool versions the project needs | Same versions on Mac, Windows and CI | asdf, nvm |
| Task runner | Task (Taskfile) | Runs common commands from one file | Works the same on Mac and Windows | Make, just |
| CI/CD | GitHub Actions | Robots that test and build on every push | Built into GitHub, free for public repos | GitLab CI, Woodpecker |
| Linting/formatting | golangci-lint (with gofmt), Oxlint, Prettier, buf | Automatic code style and mistake checks | Defaults, no debates about style; Oxlint is Vite's default (ADR 0019) | ESLint, Biome |
| Commit checks | Lefthook | Runs the lint checks before each commit | One small config, works on Mac, WSL2 and Linux (ADR 0014) | pre-commit, Husky |
| Go HTTP and logging | Go standard library (`net/http`, `log/slog`) | Built-in HTTP routing and structured logging | Fewer dependencies, fully standard | chi, zap |
| Database access | pgx + sqlc | The standard Postgres driver for Go; sqlc generates Go functions from SQL queries | Fast, maintained; you write and read real SQL (ADR 0012) | database/sql, GORM, Ent |
| Protobuf tooling | buf | Lint, breaking-change checks and code generation | Catches API-breaking changes; generators run locally (ADR 0017) | protoc |
| Styling | Tailwind CSS | Small utility classes written in the markup | Owner's choice: popular, styles next to the markup (ADR 0013) | CSS Modules, plain CSS |
| Frontend tests | Vitest + Testing Library | Frontend unit and component tests | Vite's own test runner, fast, standard | Jest |
| Integration tests | Testcontainers | Starts a real Postgres in Docker for tests | Tests the real SQL and migrations (ADR 0023) | A test database in Compose |
| Password hashing | argon2id (`golang.org/x/crypto`) | A slow, memory-hungry one-way hash for passwords | OWASP's first choice (ADR 0025) | bcrypt, scrypt |
| Rate limiting | `golang.org/x/time/rate`, in memory | Caps login and sign-up tries | Simplest; counts are short-lived (ADR 0027) | Valkey, Postgres, a Caddy plugin |
| Web routing | React Router (declarative mode) | One address per page: `/login`, `/account` | Owner's choice: real addresses, Back button works (ADR 0026) | Plain React state, TanStack Router |
| WebSocket in Go | coder/websocket | Keeps the open connections to browsers in `realtime` | Maintained, `context` support, checks the Origin header (ADR 0033) | gorilla/websocket, x/net/websocket |
| End-to-end tests | Playwright | Tests full user flows in a real browser | Two users at once in one test (ADR 0034) | Cypress |

**Planned for later** (explained in full when we get there)

| Area | Choice | Why later |
| --- | --- | --- |
| File storage | Garage or SeaweedFS (S3-compatible) | Needed with the file sharing milestone |
| Calls | WebRTC + coturn; later LiveKit or a Pion SFU for groups | Needed with the calls milestone |
| Observability | OpenTelemetry, Prometheus, Loki, Grafana | Once there are several services to watch |
| Encryption | TLS, then end-to-end with libsodium primitives | After chat works; required before offline mode |
| Native apps | Flutter (maybe) or Tauri; SQLite locally | Phase 2; Flutter only for native, React stays for web |
| Event streaming | Kafka | After NATS, as a learning module (event history, analytics) |
| Releases | release-please + GitHub Container Registry | Milestone 4: versions, changelog and tagged images |
| Orchestration | Kubernetes (k3s) + Kustomize + Argo CD | Milestone 12, first on the owner's computer |
| Server setup | Ansible (+ OpenTofu if the provider supports it) | When the owner sets up a VPS |
| Secrets in git | SOPS + age | When real deployment secrets appear |
| Load testing | k6 | Once there is something worth stressing |

## Design for the future

Kubernetes, Kafka, offline sync and end-to-end encryption come later, but these rules apply from day one so adding them never needs a rewrite.

**Ready for Kubernetes**

- Every service is **stateless**: no important data lives in memory. Data goes to Postgres, NATS or file storage. So 1 or 10 copies of a service can run.
- **Configuration comes from environment variables**, with sensible defaults. No config baked into code.
- Every service has a **health check** endpoint (`/healthz`) and logs to standard output.
- Every service has its **own Dockerfile** and can start and stop cleanly (graceful shutdown).
- Each Compose service will map to one Kubernetes deployment later.

**Ready for Kafka**

- Services talk through **events** on NATS (for example `chat.message.created`). Event names and payloads are defined in Protobuf.
- Messages are an **append-only event log**: an edit or delete is a new event, never an overwrite.
- The code that publishes events sits behind a small interface, so Kafka can replace or join NATS later.

**Ready for offline sync**

- **IDs are created by the client** as UUIDv7, so the same message sent twice is stored once (idempotency).
- The server gives each event a **sequence number per conversation**. A client syncs by asking "everything after number N".
- The data model supports **multiple devices per user** from the start.

**Ready for end-to-end encryption**

- The server should never need to read message content to work. Routing, ordering and sync use metadata only, so content can become encrypted later.

**Microservices**

- The owner wants real microservices. Each service is a separate Go program in its own container, started together by Docker Compose.
- Start with **few, meaningful services**. Split further only when there is a clear reason.
- **Each service owns its data**: its own database (or schema) in the shared Postgres container. No service reads another service's tables.
- Services talk through NATS events, or direct Connect RPC calls when an immediate answer is needed.
- Shared code (generated Protobuf, logging, config helpers) lives in a shared Go package, kept small.

## Repository structure

One monorepo named `reach`, organized by purpose. Folders that are not needed yet are created only when their milestone starts, so the tree never holds empty or unexplained parts.

```
reach/
├── README.md              # what Reach is, how to run it in 3 commands
├── CLAUDE.md              # imports AGENTS.md for Claude Code
├── AGENTS.md              # all the rules for agents (read at every start)
├── Taskfile.yml           # all common commands: task dev, task test, ...
├── .mise.toml             # exact tool versions (Go, Node, Task, buf, ...)
├── compose.yaml           # starts the whole system with Docker Compose
├── .env.example           # example settings; real .env is git-ignored
├── .gitignore
├── .gitattributes         # consistent line endings on Mac and Windows
├── .editorconfig          # basic editor settings (indent, newline)
├── apps/
│   └── web/               # React + TypeScript + Vite web app
├── services/
│   ├── hello/             # Go service: small example from milestone 1 (ADR 0011)
│   ├── auth/              # Go service: accounts, login, sessions
│   ├── chat/              # Go service: conversations, messages, events
│   └── realtime/          # Go service: WebSocket connections, pushes events
├── packages/
│   └── go/                # small shared Go code: settings, health check, database, server (ADR 0035)
├── proto/                 # Protobuf definitions: the single API contract
├── gateway/
│   └── Caddyfile          # routes browser requests to services
├── postgres/
│   └── setup.sql          # one database and one user per service (ADR 0024)
├── infra/                 # later: ansible/, k8s/, ...
├── docs/
│   ├── PROJECT_BRIEF.md   # this document
│   ├── decisions/         # decision records (ADRs)
│   ├── design/            # design docs: one per milestone, before the code
│   ├── sessions/          # summaries of chats with agents
│   ├── reviews/           # code review notes and feedback
│   ├── learn/             # learning notes per concept and milestone
│   ├── milestones/        # plan and status of each milestone
│   └── style/             # the owner's habits and patterns
└── .github/
    └── workflows/         # CI: lint, test, build
```

The exact service list may change; the services above are the starting proposal and are created in milestone 1 or later, not during initialization.

## Development environment

The owner works on both a Mac and a Windows PC, and everything must work the same on both. The Mac is also used later for iOS and macOS builds.

- **Windows:** code lives inside WSL2 (a real Linux inside Windows). Docker Desktop uses the WSL2 backend. The terminal and tools run in WSL2.
- **Mac:** Docker Desktop or OrbStack.
- **Editor:** VS Code with Claude Code. Recommended extensions are listed in `.vscode/extensions.json`.
- **Tool versions:** installed by mise from `.mise.toml`, so both machines and CI use identical versions.
- **Commands:** always through Task (`task dev`, `task test`, `task lint`). No Mac-only or Windows-only scripts.
- **Line endings:** `.gitattributes` forces LF, so Mac and Windows don't create noisy diffs.
- **Setup goal:** on a fresh machine, the owner installs Git, Docker and mise, clones the repo, and runs `task setup` then `task dev`. Nothing else.

## Milestones

There is no timeline; each milestone ends with something that runs and with its learning notes written. The order can change.

| # | Milestone | Main concepts learned |
| --- | --- | --- |
| 0 | Initialization: repo scaffold, tooling, docs structure | Git, monorepo, tooling, docs-as-code |
| 1 | Hello world end to end: React page → Caddy → Go service → Postgres, all in Docker Compose | Frontend, backend, API, containers, database |
| 2 | Accounts: sign up, login, sessions, devices | Password hashing (argon2id), tokens, security basics |
| 3 | Real-time chat: 1:1 and groups, delivered and read status | WebSocket, events, NATS, message ordering |
| 4 | First release, on the owner's computer: version `v0.1.0`, a changelog, tagged images built by CI, and a production-like Compose setup next to the dev one | Semantic versioning, release strategy, changelogs, image registries, dev vs production config |
| 5 | File sharing | Uploads, object storage |
| 6 | Observability | Logs, metrics, traces, dashboards |
| 7 | 1:1 voice and video calls | WebRTC, signaling, NAT traversal (coturn) |
| 8 | End-to-end encryption | Applied cryptography, keys, device verification |
| 9 | Native apps (desktop, then mobile) with local copy and sync | Native packaging, local-first, offline sync |
| 10 | Offline LAN mode, then Bluetooth | Peer-to-peer networking, device discovery, native platform APIs |
| 11 | Server deployment on a VPS | Linux servers, Ansible, HTTPS, backups, CD |
| 12 | Kubernetes and Kafka, first with k3s on the owner's computer | Orchestration, several copies of a service, GitOps, event streaming |
| 13 | Group calls and app releases | SFU media servers, signing, app stores, auto-update |

Milestone 4 brings releases forward, so every later milestone ends with a new version (ADR 0028). When the owner rents a VPS, milestone 11 makes a release reachable for friends.

The status of each milestone is in [milestones/README.md](milestones/README.md).

## Future considerations

These are ideas only, not part of any milestone yet. An agent must not build them unless the owner asks; it may remind the owner when one fits the current milestone.

**Feature ideas**

| Idea | What it teaches | Could fit in |
| --- | --- | --- |
| Invite-only sign-up (invite links, QR codes) | Tokens, expiry, abuse prevention | Any time (sign-up is open for now, ADR 0021) |
| Replies, reactions, edits, deletes | The event model in practice | After milestone 3 |
| Typing indicators and online status | Short-lived data, TTLs | After milestone 3 |
| Web push notifications (Web Push standard) | Service workers, browser standards | Milestones 3–5 |
| Message search | Full-text search, later a search index fed by events | Milestone 5+ |
| Link previews | Server-side request forgery (SSRF) protection | Milestone 5 |
| Voice messages | Recording in the browser, audio formats | Milestone 5 |
| Screen sharing | An extra WebRTC stream | Milestone 7 |
| Disappearing messages | Background jobs, scheduled deletion | After milestone 3 |
| Link a second device by QR code | Multi-device, sharing keys | Milestone 8 |
| Export my data, delete my account | Privacy by design | Any time |
| Installable web app (PWA), dark mode, accessibility | Web standards, a11y | Milestones 1–3 |

**Concepts to practice on purpose**

Already practiced: the transactional outbox (ADR 0031), rate limiting (ADR 0027), and testing layers: unit, integration and end-to-end (ADR 0023, 0034).

- **Resilience:** timeouts, retries with backoff, graceful degradation. Stop a container on purpose and observe.
- **API evolution:** change Protobuf messages without breaking older clients.
- **Supply-chain security:** automatic dependency updates, vulnerability scanning, signed container images.
- **Profiling and load testing:** measure before optimizing.
- **Backups and restore drills:** a backup only counts once a restore has been tested.
- **Automated releases:** Conventional Commits generate versions and the changelog (milestone 4).

**Additional tech to consider**

| Tool | What it is | Why it could fit | When | Alternatives |
| --- | --- | --- | --- | --- |
| TanStack Query | Fetches, caches and refreshes server data in React | Removes lots of hand-written loading and caching code | Later | SWR, plain fetch |
| Mailpit | A fake email server with a web inbox, for development | Test invite and password-reset emails locally | When emails are needed | MailHog |
| Valkey | In-memory key-value store (open-source Redis fork) | Rate limiting, caching, presence, if NATS KV is not enough | Milestone 3+ | Redis, NATS KV |
| Renovate | Opens pull requests to update dependencies | Keeps everything up to date automatically | Later (not wanted yet) | Dependabot |
| Trivy, govulncheck | Scan images and Go code for known vulnerabilities | Security basics for a public repo | Milestone 2+ | Grype, Snyk |
| cosign | Signs container images | Proves images were built by your CI | Milestone 4 or 11 | Notation |
| Zot | Self-hosted container image registry | Fully self-hosted images | Milestone 11 | Harbor, Docker Registry |
| Forgejo + Forgejo Runner | Self-hosted Git and CI | Option if you ever want to leave GitHub | Optional | Gitea, GitLab |
| Meilisearch | Simple, fast search engine | Better message search than Postgres alone | After milestone 5 | OpenSearch, Typesense |
| FFmpeg | Converts audio, video and images | Voice messages, thumbnails | Milestone 5 | libvips (images) |
| pprof | Go's built-in profiler | Find real performance bottlenecks | Milestone 6+ | Pyroscope |

## Learning map

Areas a senior engineer knows, and where Reach can teach them. Like the future considerations above, these are ideas, not plans: an agent brings an area in when a milestone fits it, explains it in `docs/learn/`, and keeps the scale tiny ([the goal](#the-goal)).

| Area | What it covers | Where in Reach |
| --- | --- | --- |
| Networking | TCP and UDP, DNS, TLS, NAT, HTTP/2, WebSocket, multicast and mDNS | Chat (3), calls (7), LAN discovery (10), the VPS (11) |
| Operating systems | Processes, threads, memory, file systems, signals, how containers work (namespaces, cgroups) | Graceful shutdown, Docker, native apps (9) |
| Hardware and native code | Bluetooth, drivers, platform APIs, battery and background limits | Native and offline (9, 10) |
| Clean code and design | SOLID, design patterns, refactoring, code smells | Every milestone, in reviews |
| Clean architecture | Layers, hexagonal ("ports and adapters"), domain-driven design basics, modular monolith vs microservices | From the `chat` service on (3) |
| Versioning and releases | Semantic versioning, changelogs, API versioning, branching (trunk-based), feature flags, canary and blue-green releases | First release (4), then every milestone |
| System design | Scaling out, caching, load balancing, the CAP theorem, back-pressure | Chat (3), Kubernetes (12) |
| Distributed systems | Idempotency, ordering, eventual consistency, the outbox pattern, sagas, retries | Chat (3) onward |
| Concurrency | Goroutines, race conditions, Go's race detector, locks vs channels | The `realtime` service (3) |
| Databases, in depth | Transactions, isolation levels, query plans (`EXPLAIN`), replication, backups | Chat (3), the VPS (11) |
| Security, in depth | Threat modeling, OWASP ASVS, supply chain (SBOM, signing), privacy (GDPR basics) | Every milestone, encryption (8) |
| Testing strategy | The test pyramid, end-to-end, contract tests, fuzzing (built into Go), load tests, chaos tests | Grows with each milestone |
| SRE and operations | SLOs, error budgets, alerts, runbooks, incidents, postmortems | Observability (6), the VPS (11) |
| Performance | Profiling, benchmarks, measuring before optimizing | Observability (6) onward |
| Frontend engineering | State management, accessibility, PWA, web performance | Chat (3), files (5) |
| Engineering process | Issues, pull request templates, code review, CODEOWNERS, estimation, roadmaps | Any time |
| Technical writing | Design docs, C4 architecture diagrams, a README for visitors | Every milestone, `docs/design/` |
| Data structures and algorithms | Complexity, choosing the right data structure | When a real case appears (sync, search) |

## Flexible and open decisions

Everything in this brief can change. These points are still open and must not be decided without the owner:

- **App name:** "Reach" for now.
- **License:** MIT (decided). It is short and easy to read, the most widely used open-source license, and lets anyone learn from and reuse the code with only a copyright notice.
- **Native app stack:** Flutter (native only, React stays for web) or Tauri (wraps the React app). Decided at phase 2 (milestone 9).
- **Exact service split:** `auth`, `chat` and `realtime` exist; `files` and `calls` come later. The `hello` service from milestone 1 stays as a small example (ADR 0011).
- **File storage:** Garage or SeaweedFS.
- **Group call server:** LiveKit, or a small SFU built with Pion for learning.
- **End-to-end encryption design:** which protocol (custom with libsodium, Signal-style, or MLS).
- **Kafka:** when, and for which use (event history, analytics).
- **VPS provider and Kubernetes timing:** when the owner is ready to deploy.

When one of these is decided, write an ADR and update this brief.
