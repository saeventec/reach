# Session: Owner's style and doc rules

- **Date:** 2026-11-03
- **Milestone:** between m03 and m04. Docs only, no code changes, nothing committed by the agent.

## Goal

Learn the owner's habits from their earlier projects, compare them with Reach, and make the docs cheaper for agents to use.

## What was done

- New [docs/style/](../style/README.md): the owner's habits and patterns, plus [ideas for Reach](../style/ideas-for-reach.md) (not decided).
- Golden rules 19 (readable beats short) and 20 (token-efficient docs), [ADR 0036](../decisions/0036-readable-code-and-token-efficient-docs.md).
- Docs are read on demand: agents start with `AGENTS.md` only. The rules live only there; the brief lost its duplicated sections (about 6,100 → 4,500 words).
- Every docs folder `README.md` is now an index. Outdated statuses fixed (root README, brief, milestones). One broken link fixed.

## Decisions

- Rule and workflow numbers stay stable, because other docs cite them.
- Existing code and long learning notes are not changed now; both are future steps.

## Next step

The design doc for milestone 4 (first release). Possibly first: split the `main.go` files into named steps (idea 8 in [ideas for Reach](../style/ideas-for-reach.md)).
