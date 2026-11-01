# 0034. Use Playwright for end-to-end tests

- **Status:** Accepted (the owner's choice)
- **Date:** 2026-11-01

## Context

Real-time chat needs two people at once: alice sends, bob sees it live, alice sees "read". Unit and integration tests can't check the whole path through the browser, the gateway, `chat`, NATS and `realtime`.

## Decision

Add [Playwright](https://playwright.dev) end-to-end tests that run in real browsers against the system started by Docker Compose. A task runs them locally, and CI runs them too.

## Why

- Two browser contexts in one test (alice and bob) make real-time behavior easy to test.
- Tests the Chrome, Firefox and Safari engines, maintained by Microsoft, widely used.
- Already listed in the brief for milestone 3.

## Alternatives considered

- **Cypress:** popular, but one browser tab per test makes two users at once awkward.
- **Testing by hand:** no setup, but slow and easy to forget.

## Consequences

- One new tool, browsers to download, and a slower CI job. We start with one browser engine (Chromium) and a few key tests.
