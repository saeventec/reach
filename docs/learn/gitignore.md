# .gitignore

## What it is

A file listing the files and folders git should **never** save. Each line is a pattern. Git ignores anything that matches.

## Why we use it

- **Secrets:** the repo is public, so `.env` (real passwords and keys) must never be committed.
- **Noise:** files your computer or editor makes (like `.DS_Store` on Mac) don't belong to the project.
- **Generated files:** build output and downloaded packages can be recreated, so we don't store them.

## Tiny example

```gitignore
# Ignore this exact file
.env

# Ignore this folder, wherever it is
node_modules/

# Ignore everything inside .vscode, except one file ("!" means "don't ignore")
.vscode/*
!.vscode/extensions.json
```

Lines starting with `#` are comments. A comment must be on its own line: `.env # secret` would be read as a file named ".env # secret".

Check why a file is ignored: `git check-ignore -v <file>`.

## Alternatives

- **A global gitignore** (`~/.config/git/ignore`): for your personal files, on your machine only. The project file is shared with everyone.

## Learn more

- https://git-scm.com/docs/gitignore
