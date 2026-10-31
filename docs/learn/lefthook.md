# Git hooks and Lefthook

## What it is

A **git hook** is a script git runs automatically at certain moments. The **pre-commit** hook runs when you type `git commit`, before the commit is made. If the script fails, there is no commit.

**Lefthook** manages git hooks from one file in the repo, `lefthook.yml`. Hook scripts themselves live in `.git/hooks/`, which git never shares, so everyone runs `lefthook install` once (it is part of `task setup`).

## Why we use it

Format and lint mistakes are found on your computer in a few seconds, before they reach GitHub and fail in CI. See [ADR 0014](../decisions/0014-use-lefthook-for-commit-checks.md).

## Our setup

`lefthook.yml`, shortened:

```yaml
rc: ./.lefthookrc            # load this file first (see below)

pre-commit:
  jobs:
    - name: go
      glob: "services/**/*.go"   # only when the commit has Go files
      run: task lint:go
    - name: web
      glob:
        - "apps/web/*"
        - "apps/web/**/*"
      run: task lint:web
    - name: proto
      glob: "proto/**/*.proto"
      run: task lint:proto
```

- **glob:** a job runs only if the commit contains matching files. A docs-only commit skips all three.
- **Careful with `**`:** in Lefthook it means "one or more folders", so `apps/web/**/*` does not match `apps/web/package.json`. That's why the web job has two patterns.
- **rc file:** editors like VS Code can run git without your shell's setup, so tools like `task` and `npm` would not be found. `.lefthookrc` adds mise's **shims** folder to the `PATH`. A shim is a small stand-in for a tool that runs the version from `.mise.toml`.
- **Why `./` in `./.lefthookrc`:** the hook is a `sh` script. On Linux and WSL2, `sh` is usually **dash**, a small, strict shell. Its `.` command ("run this file in the current shell") only searches the `PATH` folders for a bare name like `.lefthookrc`, not the current folder. `./` says "the file right here", which works in every shell. After changing `rc`, run `lefthook install` to rewrite the hook.
- The jobs check the whole folder as it is on disk, not only the staged part. Simple, and good enough for a small project.

What a blocked commit looks like:

```
┃  web ❯
[warn] src/App.tsx
[warn] Code style issues found in the above file. Run Prettier with --write to fix.
```

Fix it with `task format`, then `git add` and commit again.

## Skipping the hook

`git commit --no-verify` skips the hooks. Use it only in an emergency: CI runs the same checks anyway (see [ci.md](ci.md)).

## Alternatives

- **pre-commit:** the most popular hook manager, but needs Python.
- **Husky:** popular in JavaScript projects, needs npm at the repo root.
- **Plain `.git/hooks` scripts:** no tool, but not shared through git.

## Learn more

- https://lefthook.dev
- https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks
- https://mise.jdx.dev/dev-tools/shims.html
