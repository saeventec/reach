# Style

The owner's own habits, patterns and standards: how they like software to be built. They were learned from the owner's earlier projects, before Reach.

This folder is a reference for **future steps**, not a rule list. The current code doesn't have to follow it. When we plan a new milestone or a refactor, we read these notes and decide which ideas to use. A decision to adopt one becomes an ADR in `docs/decisions/`.

## The core idea

> **Simple, clean, organized. Predictable beats clever.**

When every part has the same shape, you learn one part and you know them all. Almost every habit below comes back to this.

## Notes

| Note | What it covers |
| --- | --- |
| [Organization](organization.md) | Folders, files, and repeating one shape everywhere |
| [Design](design.md) | Layers, small interfaces, wiring, swappable parts |
| [Code style](code-style.md) | Names, functions, control flow, comments |
| [Errors and responses](errors-and-responses.md) | One error vocabulary, one response shape |
| [Configuration](configuration.md) | Settings from the environment, read once, in one place |
| [Testing](testing.md) | Fakes, test helpers, readable checks |
| [Delivery and operations](delivery-and-operations.md) | Containers, environments, automation, deployment |
| [Repository hygiene](repository-hygiene.md) | The files and habits every repository starts with |
| [Ideas for Reach](ideas-for-reach.md) | Where Reach differs today, and what we might change later |

## How to add a note

1. Copy `_template.md`.
2. Name it after the area, in lowercase: `logging.md`.
3. One section per habit: what it is, why it helps, a tiny example.
4. Keep it general: no project names, and no language unless the habit only makes sense in one.
5. Add it to the table above.
