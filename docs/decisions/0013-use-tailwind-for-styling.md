# 0013. Use Tailwind CSS for styling

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

The web app needs a way to style its components.

## Decision

Use Tailwind CSS: you style elements with small class names written directly in the markup, such as `p-4 text-lg`.

## Why

- The owner's choice: very popular, with lots of learning material.
- Styles sit next to the markup they change, so one file shows the whole component.
- Only the classes we use end up in the final CSS file, so it stays small.

## Alternatives considered

- **CSS Modules:** plain CSS scoped to one component, built into Vite, no new tool.
- **Plain global CSS:** simplest, but styles can clash as the app grows.
- **vanilla-extract:** styles written in TypeScript, less common.

## Consequences

One more tool and its syntax to learn. It is set up in step 5 of milestone 1, following the official Tailwind guide for Vite at that time.
