# Design docs

One design doc (also called an RFC, "Request for Comments") per milestone, written before any code: the problem, the options, the trade-offs and a diagram, so ideas can be questioned while changing them is still cheap.

- A **design doc** looks at the whole milestone.
- An **ADR** in [decisions/](../decisions/README.md) records one single choice that comes out of it.
- The **milestone file** in [milestones/](../milestones/README.md) is the step list and its status.

## Index

| Design doc | Status |
| --- | --- |
| [m03 Real-time chat](m03-real-time-chat.md) | Built (2026-11-01) |

Milestones 0 to 2 started before design docs were introduced.

## How to add one

1. Copy `_template.md`, name it like its milestone file: `m04-first-release.md`.
2. The owner reads it and asks questions before the code starts. Keep it short.
3. If it changes while building, write the change and the reason in it.
4. Add it to the index.
