# Session: Learning goals and showcase

- **Date:** 2026-11-01
- **Milestone:** between m02 and m03

## Goal

Make the project's goal clearer before milestone 3, and update the docs so every later session follows it. No code changes.

## What was done

- The brief's goal now says: learn what a senior engineer knows, as a public showcase, with "real concepts, tiny scale", and "observe first, direct later".
- Golden rule 18 (the agent is a senior mentor) and workflow rule 11 (use the learning map), in the brief and in [AGENTS.md](../../AGENTS.md).
- A new milestone 4, "first release", on the owner's computer. The later milestones were renumbered (4 to 12 became 5 to 13), also in the ADRs, learning notes, the milestone 2 session summary and the Caddyfile comment.
- A [learning map](../PROJECT_BRIEF.md#learning-map) in the brief: networking, operating systems, clean architecture, versioning and more, each with where Reach can teach it.
- A new folder [docs/design/](../design/) for one design doc per milestone, with a README and a template.
- Cleanup: the finished "Initialization task" section left the brief (it lives in [m00](../milestones/m00-initialization.md)), and release-please moved from "ideas" to "planned".

## What was explained

- Why being senior is about judgment and trade-offs, not the number of tools.
- "Real concepts, tiny scale": a scale model of enterprise systems, and labeled learning modules.
- Design docs (RFCs) vs ADRs vs milestone files.

## Decisions

- [0028](../decisions/0028-reach-is-a-learning-showcase.md): Reach is a learning showcase; first release as milestone 4 (option A: on the owner's computer, the VPS stays milestone 11); renumber the milestones; design docs per milestone. All chosen by the owner.
- The owner doesn't write code for now, and directs later. The new learning areas are ideas for the future, not plans.

## Open questions

- Milestone 4 details (release-please or another tool, image tags, the production Compose setup): decided in its design doc.

## Next step

Plan [milestone 3: real-time chat](../PROJECT_BRIEF.md#milestones). Start with its design doc `docs/design/m03-real-time-chat.md`, then the milestone file. Decide there: the `chat` and `realtime` services, the code structure of `chat` (for example hexagonal architecture), sharing Go code in `packages/go/`, and Playwright for end-to-end tests. Still open from milestone 2: check it on the Mac.
