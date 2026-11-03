# Repository hygiene

The files and habits every repository starts with.

## The starter files

- **What:** every repository starts with:
  - an editor config file (encoding, line endings, indentation, line length)
  - a git-ignore file (local secrets, editor and system files)
  - a `.env.example` with every setting and safe sample values
  - a compose file for supporting services
  - a container build file
  - a command runner file (common commands with short names)
  - a CI workflow
  - a README
- **Why:** the basics are never forgotten, and every project feels familiar.

## A command runner with built-in help

- **What:** common commands (`test`, `up`, `down`, `build`, `serve`, `sync`) have short names in one file, with a `help` command that lists them with one-line descriptions.
- **Why:** nobody memorizes long commands, and the help is the documentation.
- **Example:**

```
Usage: run <command>
  test     Run unit tests
  up       Start the supporting services
  down     Stop the supporting services
  build    Build the image
```

## Settings that can be overridden

- **What:** values in the command runner (image name, tag, account) are variables with defaults that a caller can override.
- **Why:** the same commands work for different accounts and versions.

## A short, practical README

- **What:** the README says what the project is, how to start it (a few bullet steps), its commands, where to reach it, a progress checklist, reading links, and the author.
- **Why:** a newcomer can start in minutes, and the checklist shows the progress at a glance.
- **Example:**

```
## Develop
- Install the command runner and containers
- Copy .env.example to .env
- Run "help"

## Progress
- [x] Config   - [x] Logging   - [x] Database   - [x] Auth   - [x] CI
```

## A progress checklist for the big pieces

- **What:** the README lists the main capabilities (config, logging, database, cache, migrations, auth, monitoring, containers, tests, CI, API docs) as checkboxes.
- **Why:** a clear plan, and a satisfying record of what's done.

## Learn from good sources, and link them

- **What:** the README links to the guides the project follows: language guides, clean architecture, the Twelve-Factor App, standard project layouts, and the official docs of the main libraries.
- **Why:** shows where the ideas come from, and gives readers the next step.

## Ship a ready-to-use API collection

- **What:** the repository includes an API client collection (requests grouped by role and feature) and an environment file with the base address.
- **Why:** anyone can try every endpoint in a minute.

## CI that builds the artifact

- **What:** a workflow with clearly named steps (check out, set up, log in, build and push) runs on every push to the main branch. Credentials come from the CI's secret store.
- **Why:** every merge produces a ready image, with no manual build.
