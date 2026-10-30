# Reach

Reach is a communication app: chat (1:1 and groups), voice and video calls, and file sharing. It starts as a web app that runs fully on your own computer, and grows into native apps that also work offline.

Reach is built from scratch as a learning project, to practice professional software engineering end to end: backend, frontend, databases, security, messaging, containers, CI/CD and more. The full plan is in the [project brief](docs/PROJECT_BRIEF.md).

## Status

Milestone 0: the repository skeleton. There is no application code yet. See [docs/milestones/](docs/milestones/).

## Set up

You need three tools: [Git](https://git-scm.com/downloads), [Docker Desktop](https://docs.docker.com/get-started/get-docker/), and [mise](https://mise.jdx.dev/getting-started.html). On Windows, work inside WSL2.

```sh
git clone <this repository's url>
cd reach
mise install   # installs Go, Node.js and Task, in the versions from .mise.toml
task setup     # prepares the project
```

Run `task` to see all available commands.

## Documentation

- [Project brief](docs/PROJECT_BRIEF.md): goals, architecture, tech stack, milestones
- [Rules for agents](AGENTS.md): how code is written and how work is done
- [Decisions](docs/decisions/), [learning notes](docs/learn/), [milestones](docs/milestones/), [sessions](docs/sessions/), [reviews](docs/reviews/)

## License

[MIT](LICENSE)
