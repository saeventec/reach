# Session: Project initialization

- **Date:** 2026-10-31
- **Milestone:** m00

## Goal

Start milestone 0: check the environment, then create the repository skeleton described in the brief.

## What was done

- Checked the Mac environment read-only. Found up to date: Git, Docker Desktop and Compose, Go, gh, VS Code. Missing: mise (the owner installed it with Homebrew). Task and the project's Node.js version come from mise.
- Created the skeleton: rules files, `docs/` structure, decision records 0001 to 0010, learning notes, milestone files, and tooling config.
- The owner activated mise in `~/.zshrc` and ran `mise install`. Inside the project, Go, Node.js and Task now come from mise.
- The owner reviewed all files: no changes requested.
- Created the public GitHub repo `saeventec/reach` (with a description and topics) and pushed `main`.
- Brief changes: replaced the lost architecture diagram embed with a text diagram, and renamed this session file to today's date.

## What was explained

- Each tool in the environment check, and what "LTS" means.
- [Git](../learn/git.md), [monorepo](../learn/monorepo.md), [Markdown](../learn/markdown.md), [mise](../learn/mise.md), [Task](../learn/task.md), [.gitignore](../learn/gitignore.md), [.gitattributes](../learn/gitattributes.md), [.editorconfig](../learn/editorconfig.md).

## Decisions

- Tool versions in `.mise.toml`: Go 1.27.3, Node.js 26.13.0 (LTS, the stable line rather than the newest), Task 3.54.0.
- VS Code: recommend Microsoft's "Container Tools" extension, which replaced the older "Docker" extension.
- The repo lives on the owner's personal GitHub account, `saeventec`.
- The `main` branch is set only for this repo, not in the global git config.

## Problems solved

- The mise activation line in `~/.zshrc` pointed to `~/.local/bin/mise` (the install-script path), but Homebrew installs mise elsewhere. Fixed by using `eval "$(mise activate zsh)"` without a path. Check with `mise doctor`.

## Open questions

- Run the Windows (WSL2) check: clone, `mise install`, `task setup`.
- Optional: remove the extra Homebrew copy of Task (`brew uninstall go-task`); mise provides it.

## Next step

Check the setup on Windows (WSL2), then start [milestone 1](../milestones/m01-hello-world.md).
