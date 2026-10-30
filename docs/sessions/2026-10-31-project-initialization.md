# Session: Project initialization

- **Date:** 2026-10-31
- **Milestone:** m00

## Goal

Start milestone 0: check the environment, then create the repository skeleton described in the brief.

## What was done

- Checked the Mac environment read-only. Found up to date: Git, Docker Desktop and Compose, Go, gh, VS Code. Missing: mise (the owner installed it with Homebrew). Task and the project's Node.js version come from mise.
- Created the skeleton: rules files, `docs/` structure, decision records 0001 to 0010, learning notes, milestone files, and tooling config.
- Brief changes: replaced the lost architecture diagram embed with a text diagram, and renamed this session file to today's date.

## What was explained

- Each tool in the environment check, and what "LTS" means.
- [Git](../learn/git.md), [monorepo](../learn/monorepo.md), [Markdown](../learn/markdown.md), [mise](../learn/mise.md), [Task](../learn/task.md), [.gitignore](../learn/gitignore.md), [.gitattributes](../learn/gitattributes.md), [.editorconfig](../learn/editorconfig.md).

## Decisions

- Tool versions in `.mise.toml`: Go 1.27.3, Node.js 26.13.0 (LTS, the stable line rather than the newest), Task 3.54.0.
- VS Code: recommend Microsoft's "Container Tools" extension, which replaced the older "Docker" extension.
- The owner creates the GitHub repo and pushes, after reviewing.
- The `main` branch is set only for this repo, not in the global git config.

## Open questions

- mise must be activated in the owner's shell (`~/.zshrc`) by the owner.

## Next step

The owner reviews every file and asks questions (saved in `docs/reviews/`), then creates the GitHub repo and pushes. Then start milestone 1.
