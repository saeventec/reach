# Reach

[![CI](https://github.com/saeventec/reach/actions/workflows/ci.yml/badge.svg)](https://github.com/saeventec/reach/actions/workflows/ci.yml)

Reach is a communication app: chat (1:1 and groups), voice and video calls, and file sharing. It starts as a web app that runs fully on your own computer, and grows into native apps that also work offline.

Reach is built from scratch as a learning project, to practice professional software engineering end to end: backend, frontend, databases, security, messaging, containers, CI/CD and more. The full plan is in the [project brief](docs/PROJECT_BRIEF.md).

## Status

Milestone 1 done: a React page shows a greeting read from Postgres, through the Caddy gateway, all started with `task dev`. See [docs/milestones/](docs/milestones/).

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
- [Decisions](docs/decisions/), [learning notes](docs/learn/), [milestones](docs/milestones/), [sessions](docs/sessions/), [reviews](docs/reviews/)

## License

[MIT](LICENSE)
