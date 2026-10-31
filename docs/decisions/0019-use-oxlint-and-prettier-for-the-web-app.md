# 0019. Use Oxlint and Prettier for the web app

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

The web app needs a linter (to find likely bugs) and a formatter (one standard style). The brief planned ESLint and Prettier, but Vite's React starter now uses Oxlint by default.

## Decision

Use Oxlint as the linter, with the config from Vite's starter, and Prettier as the formatter, with its defaults.

## Why

- Oxlint: Vite's default today, one package, almost no config, very fast. It has the important React rules built in (rules of hooks).
- Prettier: the standard formatter for web code. Oxc's own formatter, Oxfmt, is not yet at version 1.
- Both follow Golden rule 12: use defaults.

## Alternatives considered

- **ESLint:** the long-time standard with the most rules and plugins, but 6 packages and a config file for our setup, and much slower.
- **Biome:** one tool for linting and formatting, fast, fewer rules and a smaller community.
- **Oxfmt:** Oxc's formatter, not yet stable.

## Consequences

If we ever need a rule only ESLint has, we can run ESLint next to Oxlint. When Oxfmt reaches version 1, we can reconsider Prettier.
