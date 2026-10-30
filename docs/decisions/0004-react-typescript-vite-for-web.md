# 0004. React, TypeScript and Vite for the web app

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The first client is a web app. We need a UI library, a language and a build tool.

## Decision

Build the web app with React (UI), TypeScript (language) and Vite (dev server and bundler).

## Why

- React is the most used UI library, with the most learning material.
- TypeScript catches mistakes before the code runs.
- Vite needs almost no configuration and is very fast.

## Alternatives considered

- **Svelte, Vue or Solid:** simpler or faster in places, but less learning material and fewer jobs.
- **Plain JavaScript:** no types, so more mistakes are found late.
- **Next.js:** adds a server framework we don't need, since our backend is Go.

## Consequences

The native apps (phase 2) may use a different stack (Flutter or Tauri). The web app stays React.
