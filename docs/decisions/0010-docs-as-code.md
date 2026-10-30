# 0010. Docs as code in docs/

- **Status:** Accepted
- **Date:** 2026-10-30

## Context

The owner works on two computers with different AI agents. All context (decisions, sessions, reviews, learning) must move with the code.

## Decision

Write all documentation as plain Markdown files in `docs/`, committed to git next to the code. `AGENTS.md` holds the rules for agents, and `CLAUDE.md` imports it.

## Why

- Docs travel with the code, on every device and to every agent.
- Changes to docs are reviewed and versioned like code.
- Markdown is simple, readable as plain text, and shown nicely on GitHub.

## Alternatives considered

- **A wiki (GitHub Wiki, Notion):** easy to edit, but separate from the code and easy to forget.
- **A docs website (Docusaurus, MkDocs):** nicer to browse, but needs a build step we don't need yet.

## Consequences

Docs must be updated in the same commit as the change they describe. The repo is public, so docs never contain secrets or private data.
