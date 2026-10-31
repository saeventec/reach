# mise

## What it is

mise is a tool version manager. It reads `.mise.toml` and installs the exact versions of the tools the project needs (Go, Node.js, Task, buf, golangci-lint, Lefthook, ...). When you are inside the project folder, those versions are the ones you use.

Analogy: a recipe that lists exact ingredients, so the dish tastes the same in every kitchen (your Mac, your Windows PC, and CI).

## Why we use it

Both of your computers and CI get identical tool versions from one file. See [ADR 0009](../decisions/0009-use-mise-and-task-for-tooling.md).

## Tiny example

Our `.mise.toml`:

```toml
[tools]
go = "1.27.3"
node = "26.13.0"
task = "3.54.0"
```

`[tools]` is a section header (TOML format). Each line below it is `tool = "version"`.

```sh
mise install     # install the versions listed in .mise.toml
mise ls          # show which versions are installed and active
go version       # inside the project: prints go1.27.3
```

## One-time setup

mise must be **activated** in your shell, so it can switch versions when you enter a folder. Add this line to the end of `~/.zshrc` (Mac) or `~/.bashrc` (WSL2), then open a new terminal:

```sh
eval "$(mise activate zsh)"   # on WSL2 with bash: mise activate bash
```

If you installed mise with Homebrew, use the line exactly as shown, without a path. A line like `eval "$(~/.local/bin/mise activate zsh)"` only works when mise was installed with its own install script. Run `mise doctor` to check: it should say `activated: yes`.

The first time mise sees a project's config file, it may ask you to **trust** it (`mise trust`). This is a safety check: config files can run commands, so mise only reads files you approved.

## Alternatives

- **asdf:** the older tool mise is inspired by; slower.
- **nvm:** only manages Node.js versions.
- **Homebrew:** installs one global version per tool, not per project.

## Learn more

- https://mise.jdx.dev
