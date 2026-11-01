# CI with GitHub Actions

## What it is

**CI** ("continuous integration") means: every change is checked automatically on a clean computer, not only on yours. If a check fails, everyone sees it, and the change is fixed before it causes trouble.

**GitHub Actions** is GitHub's CI. You describe a **workflow** in a YAML file in `.github/workflows/`. GitHub runs it on a fresh Linux computer (a **runner**) for each push or pull request, and shows a green check or a red cross next to the commit.

- **Workflow:** one YAML file (`ci.yml`).
- **Job:** a group of steps that run on one runner (`check`).
- **Step:** one command (`run:`) or one ready-made **action** (`uses:`).
- **Action:** a reusable step published by someone, like `actions/checkout`, which downloads the repo.

## Why we use it

It is built into GitHub and free for public repos. Lefthook can be skipped, and "it works on my computer" isn't proof: CI is the final check. See [ADR 0020](../decisions/0020-ci-with-github-actions-mise-and-task.md).

## Our workflow

`.github/workflows/ci.yml` runs on every push to `main` and every pull request:

1. **Get the code** (`actions/checkout`), with the full history so buf can compare with older commits.
2. **Install the tools** from `.mise.toml` (`jdx/mise-action`): the same versions as on your computer.
3. **`task setup`**, **`task lint`**: the same commands you run.
4. **Generated code up to date:** run `task generate`, then `git diff --exit-code` fails if anything changed.
5. **Breaking API changes:** `buf breaking` compares the `.proto` files with the commit before this change.
6. **`task test`**, then **`task test:race`** (realtime with Go's race detector).
7. **End-to-end tests:** install Chromium with the Linux libraries it needs, then **`task test:e2e`** ([playwright.md](playwright.md)). If any step failed, **`actions/upload-artifact`** keeps the traces of failed end-to-end tests for 7 days: download them from the run's page ("Artifacts"), and open them with `npx playwright show-trace`.
8. **`task build`** (the web app and all container images).

Because CI uses the same Task commands, you can always reproduce a CI failure on your own computer.

## Security habits in the file

- `permissions: contents: read`: the workflow can only read the repo.
- Actions are pinned to an exact commit (`actions/checkout@3d3c42e…  # v7.0.1`). A tag like `v7` can later point to different code. A commit can't.
- `persist-credentials: false`: the GitHub token isn't left behind in the downloaded repo.
- Values from GitHub (like `${{ github.event.before }}`) go into `env:` first, never straight into a `run:` command, a habit that prevents script injection.

## Where to look

On GitHub: the repo's **Actions** tab, or the check mark next to a commit. Click a failed step to see its output. From the terminal:

```sh
gh run list              # the latest runs
gh run watch             # follow the current run live
gh run view --log-failed # show the output of the failed steps
```

## Alternatives

- **GitLab CI:** the same idea, built into GitLab.
- **Woodpecker or Forgejo Actions:** self-hosted CI, if we ever leave GitHub.
- **Jenkins:** the classic self-hosted CI server, a lot to maintain.

## Learn more

- https://docs.github.com/en/actions/get-started/understand-github-actions
- https://docs.github.com/en/actions/reference/security/secure-use
