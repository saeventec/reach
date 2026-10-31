# Milestone 0: Initialization

- **Status:** Done (Mac, and Windows WSL2 checked by the owner on 2026-10-31).

## Goal

An empty, well-organized repository skeleton: rules, docs structure and tooling. No application code.

## Steps

- [x] Check the development environment (Mac: Git, Docker, mise, gh, VS Code)
- [x] Owner installs mise and activates it in `~/.zshrc`
- [x] `mise install` and `task` work on Mac
- [x] `git init` with `main` as the default branch
- [x] `README.md`, `AGENTS.md`, `CLAUDE.md`, `LICENSE`
- [x] `docs/` folders, each with a `README.md` and `_template.md`
- [x] First decision records (0001 to 0010)
- [x] Learning notes for this step
- [x] `.gitignore`, `.gitattributes`, `.editorconfig`, `.env.example`
- [x] `.mise.toml`, `Taskfile.yml`, `.vscode/extensions.json`
- [x] Session summary
- [x] First commit: `chore: initialize repository`
- [x] Owner reviews every file ([review](../reviews/2026-10-31-review-initialization.md))
- [x] GitHub repo created and pushed: https://github.com/saeventec/reach
- [x] Check on Windows (WSL2): clone, `mise install`, `task setup` (checked at the end of milestone 1: everything worked)

## Concepts learned

- [Git](../learn/git.md), [monorepo](../learn/monorepo.md), [Markdown](../learn/markdown.md)
- [mise](../learn/mise.md), [Task](../learn/task.md)
- [.gitignore](../learn/gitignore.md), [.gitattributes](../learn/gitattributes.md), [.editorconfig](../learn/editorconfig.md)

## Done when

The owner can clone the repo on Mac or Windows, run `mise install` and `task setup`, read every file, and understand why each one exists.
