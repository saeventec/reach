# Reach

[![CI](https://github.com/saeventec/reach/actions/workflows/ci.yml/badge.svg)](https://github.com/saeventec/reach/actions/workflows/ci.yml)

Reach is a communication app: chat (1:1 and groups), voice and video calls, and file sharing. It starts as a web app that runs fully on your own computer, and grows into native apps that also work offline.

Reach is built from scratch as a learning project and a showcase. It practices what senior engineers use for enterprise-grade software, at a tiny scale that runs on one computer: backend, frontend, APIs, databases, security, networking, messaging, containers, CI/CD, observability, clean architecture, versioning, releases and more. Every decision is written down with its reasons and trade-offs. The full plan is in the [project brief](docs/PROJECT_BRIEF.md).

## Status

Milestone 2 done: accounts. You can sign up, log in, stay logged in, see the devices you are logged in on, and log out (also another device). Passwords are hashed with argon2id, sessions live in HttpOnly cookies, and logins are rate limited. See [docs/milestones/](docs/milestones/).

## Set up

You need three tools: [Git](https://git-scm.com/downloads), [Docker Desktop](https://docs.docker.com/get-started/get-docker/), and [mise](https://mise.jdx.dev/getting-started.html). On Windows, work inside WSL2. Activate mise in your shell once, as shown in [docs/learn/mise.md](docs/learn/mise.md).

```sh
git clone https://github.com/saeventec/reach.git
cd reach
mise trust     # allows mise to read this project's .mise.toml
mise install   # installs Go, Node.js, Task and the other tools, in the versions from .mise.toml
task setup     # installs the libraries, creates .env, turns on the commit checks
task dev       # builds and starts the whole system (Ctrl+C stops it)
task stop      # stops it and removes its containers; your data is kept
```

Then open http://localhost:8080.

Run `task` to see all available commands.

## Documentation

- [Project brief](docs/PROJECT_BRIEF.md): goals, architecture, tech stack, milestones
- [Rules for agents](AGENTS.md): how code is written and how work is done
- [Decisions](docs/decisions/), [design docs](docs/design/), [learning notes](docs/learn/), [milestones](docs/milestones/), [sessions](docs/sessions/), [reviews](docs/reviews/)

## License

[MIT](LICENSE)
