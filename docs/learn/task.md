# Task

## What it is

Task is a task runner: it runs common commands that are written down in `Taskfile.yml`. Instead of remembering long commands, you type `task setup` or `task dev`.

Analogy: a list of named shortcuts that everyone on the project shares.

## Why we use it

The same commands work on Mac, Windows (WSL2) and CI. See [ADR 0009](../decisions/0009-use-mise-and-task-for-tooling.md).

## Tiny example

Our `Taskfile.yml`:

```yaml
version: "3"          # the Taskfile format version

tasks:
  default:            # runs when you type just "task"
    desc: List all available tasks
    cmds:
      - task --list

  setup:
    desc: Install the tool versions from .mise.toml
    cmds:
      - mise install
```

- `desc` is a short description, shown by `task --list`.
- `cmds` is the list of shell commands the task runs, in order.

```sh
task          # lists tasks
task setup    # runs "mise install"
```

## Alternatives

- **Make:** available everywhere, but its syntax is tricky (tabs matter) and it works differently on Windows.
- **just:** a simple command runner, less common.
- **npm scripts:** only fit JavaScript projects.

## Learn more

- https://taskfile.dev
