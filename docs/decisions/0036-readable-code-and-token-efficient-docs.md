# 0036. Readable code and token-efficient docs

- **Status:** Accepted (asked for by the owner)
- **Date:** 2026-11-03

## Context

- Some code is hard to follow: long functions, and responsibilities that aren't clearly separated. The `main.go` files do too many things in one place.
- Agents read and write `docs/` constantly. Every session started by reading the whole brief (about 6,000 words), and the rules were written twice (`AGENTS.md` and the brief).

## Decision

- **Golden rule 19, readable beats short:** small blocks with one job, functions of about 30 lines at most, `main` and handlers only list named steps.
- **Golden rule 20, token-efficient docs:** one topic per file, clear headings, facts first, lists and tables, explain once and link, current indexes.
- **Docs are read on demand.** An agent starts with `AGENTS.md` only, then opens the one file and section a task needs. A full read or search happens only when needed or asked.
- **One home for the rules:** `AGENTS.md`. The brief keeps goals, architecture, stack and milestones, and links to the rules.
- **Every docs folder `README.md` is an index** (file, one-line summary or status), so a file can be found without opening others.
- **Code comments** explain *why* in a line or two and link to `docs/learn/` for concepts.

## Why

- Small named steps read like a table of contents: easier to review, test and change.
- Reading only what a task needs saves time and tokens on every session.
- Rules written once can't drift apart.
- Explaining a concept once (in `docs/learn/`) keeps it complete without repeating it.

## Alternatives considered

- **Keep rule 6 only ("simple, clean, small"):** too general; it didn't prevent long `main` files.
- **Enforce function length with a linter (golangci-lint's `funlen`):** possible later; a hard limit can push toward awkward splits, so it starts as guidance.
- **Keep reading the brief at every start:** simple, but most tasks need one section of it, or none.
- **Shorter docs by explaining less:** would break golden rules 1 to 3. We keep the explanations, but write them once.

## Consequences

- Rule numbers (1 to 20) and workflow numbers (1 to 11) stay stable, because other docs cite them.
- New code and docs follow the rules from now on. Existing code is **not** changed now; refactoring it (for example the `main.go` files) is a planned future step ([ideas for Reach](../style/ideas-for-reach.md)).
- Long learning notes and the milestone 3 design doc are read on demand, so they stay as they are for now; tightening them is a future step.
- Watch out: "small" must not become "scattered". A function is split only when the new piece has a clear name and one job.
