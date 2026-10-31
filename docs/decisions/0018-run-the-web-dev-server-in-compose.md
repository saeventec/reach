# 0018. Run the web app's development server in Compose, with Compose Watch

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

`task dev` must start the whole system with one command, including the web app. While developing, a saved file should show up in the browser within a moment.

## Decision

Run Vite's development server in a container (`apps/web/Dockerfile`), and keep it up to date with Compose Watch: changed files are synced into the container, and a changed `package-lock.json` rebuilds the image.

## Why

- One command (`task dev`) starts everything, the same way on Mac and Windows (WSL2).
- The container installs its own Linux libraries, so the Mac or Windows copy in `node_modules` never gets in the way.
- Compose Watch is Docker's current recommended way to develop in containers.

## Alternatives considered

- **Bind mount (share the folder with the container):** the older way. `node_modules` is shared too, and its Mac or Windows versions of some libraries don't run in Linux.
- **Run Vite on your computer, outside Docker:** the simplest and fastest, but then `task dev` doesn't start everything in one go.

## Consequences

The libraries are installed twice: in the container, and on your computer (by `task setup`) so the editor and `task test` work. The image is for development only. A real server (milestone 11) will serve the built files instead.
