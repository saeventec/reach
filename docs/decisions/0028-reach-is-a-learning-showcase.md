# 0028. Reach is a learning showcase: real concepts, tiny scale

- **Status:** Accepted
- **Date:** 2026-11-01

## Context

The brief said Reach is for learning software engineering "end to end". The owner made the goal clearer:

- Learn what a senior engineer knows, including what large companies use for enterprise-grade software (Kubernetes, microservices, native and offline features), and areas the brief missed: networking, operating systems, clean code, clean architecture, versioning and release strategy.
- Use the repo as a public showcase of the owner's skills.
- Keep it simple and small: learn each advanced topic at a tiny scale, but don't skip a topic only because Reach doesn't strictly need it.
- The owner observes and reviews for now, and will direct the design later.
- The agent acts as a senior mentor: it challenges ideas and always adds its own suggestion.

Releases came last (milestone 12), so versioning and release strategy would have been learned very late.

## Decision

- **Real concepts, tiny scale.** Advanced topics come in as small, planned, labeled learning modules (for example 2 or 3 copies of a service, k3s on the owner's computer). Each one's ADR says what it teaches and what it would cost in a real project.
- **A new milestone 4, "first release", on the owner's computer:** version `v0.1.0`, a changelog, tagged images built by CI on GitHub's container registry, and a production-like Compose setup. The old milestones 4 to 12 become 5 to 13. A VPS (now milestone 11) comes when the owner is ready to rent one.
- **Design docs** in `docs/design/`, one per milestone, before the code.
- **A learning map** in the brief, with areas to bring in when a milestone fits.
- **Golden rule 18:** the agent is a senior mentor, not only a builder.

## Why

- Judgment is the senior skill: an honest "added to learn X, here is the trade-off" shows it better than a list of tools.
- A small scale keeps everything on one computer and understandable, while the concepts stay real.
- Releasing early means every later milestone ends with a version, so releases are practiced many times, not once.
- Design docs are how large teams agree on a change before building it, and they make the showcase easy to follow.

## Alternatives considered

- **Keep the old order:** simpler, but versioning and releases would only be practiced at the very end.
- **Release on a VPS at once (option B):** friends could use Reach sooner, but it pulls a lot of server work forward. It stays milestone 11.
- **Add milestone 4 as "3b":** no renumbering, but an odd number that stays forever. The owner chose to renumber.
- **Design notes inside the milestone file:** fewer files, but it mixes "why and how" with the step list.

## Consequences

- Milestone numbers 4 and up changed. The docs were updated on 2026-11-01. Older commit messages still use the old numbers.
- Some extra work per milestone (a design doc), in return for clearer plans.
- Learning modules add code that a small app wouldn't need. Each one must be labeled and explained, so it never looks like accidental complexity.
