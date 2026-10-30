# Git

## What it is

Git records every change to your files, like a time machine for your project. Each saved snapshot is a **commit**, with a message that says what changed and why. GitHub is a website that hosts git repositories and runs automation (CI) on them.

Key words:

- **Repository (repo):** a project folder tracked by git. The history lives in a hidden `.git/` folder.
- **Commit:** one saved snapshot, with a message.
- **Branch:** a separate line of work. Our main branch is called `main`.
- **Staging:** choosing which changes go into the next commit (`git add`).
- **Remote:** a copy of the repo somewhere else, like GitHub. `push` sends commits there, `pull` gets them.

## Why we use it

To see the full history, undo mistakes, review changes, and work on two computers. It is the industry standard.

## Tiny example

```sh
git status                          # what changed?
git add README.md                   # stage a file for the next commit
git commit -m "docs: fix a typo"    # save a snapshot
git log --oneline                   # list past commits
```

## Commit messages

We use **Conventional Commits**: `type(scope): short description`. The type says what kind of change it is:

- `feat`: a new feature, like `feat(auth): add sign-up endpoint`
- `fix`: a bug fix
- `docs`: documentation only
- `chore`: maintenance, like `chore: initialize repository`

## Alternatives

- **Mercurial:** similar idea, rarely used today.
- **Subversion (SVN):** older, keeps history on one central server.

## Learn more

- https://git-scm.com/doc
- https://www.conventionalcommits.org
