# Monorepo

## What it is

A monorepo is one repository that holds many projects. Reach keeps the web app, all services, shared code, the API contract, infrastructure and docs in one repo, in folders by purpose.

Analogy: one big binder with tabs, instead of a separate binder for each subject.

## Why we use it

- Clone once, and you have everything.
- A change across the API, a service and the web app is one commit.
- Easy to search and understand the whole system.

See [ADR 0001](../decisions/0001-use-a-monorepo.md).

## Tiny example

```
reach/
├── apps/web/        # the web app
├── services/chat/   # one Go service
├── proto/           # the API contract both of them use
└── docs/
```

## Alternatives

- **Polyrepo:** one repo per project. Clear ownership, but changes across repos are hard to coordinate.

## Learn more

- https://monorepo.tools
