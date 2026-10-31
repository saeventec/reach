# 0026. Use React Router for the web app's pages

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

With accounts (milestone 2), the web app gets several screens: log in, sign up, and your account. The app needs a way to move between them.

## Decision

Use React Router 8, in its simplest "declarative mode": `BrowserRouter`, `Routes`, `Route`, `Link` and `Navigate`. Each screen has its own address: `/login`, `/signup`, `/account`. The owner chose this over plain React state.

## Why

- Real addresses: the Back and Forward buttons and bookmarks work.
- The most used router for React, with good documentation and many examples.
- Declarative mode is small: routes are components, with no data-loading system to learn.
- Chat (milestone 3) will add more pages, like one per conversation (`/chats/:id`), which a router handles well.

## Alternatives considered

- **Plain React state:** no library, but no addresses, and the Back button leaves the app. The agent's first suggestion, to keep this step smaller.
- **TanStack Router:** fully type-safe, but newer and more to learn.
- **wouter:** very small, but less used and less documented.

## Consequences

- One new library in the web app.
- When the built app is served by our own server later (not Vite's development server), the server must answer every page address with `index.html`.
- React Router's data mode (loaders and actions) or TanStack Query could come later, if loading data by hand gets repetitive.
