# Design docs

One design doc per milestone, written at its start, before any code. Large companies call these "design docs" or "RFCs" (Request for Comments): a short paper that says what we want to build and how, so people can question it while changing it is still cheap.

How it differs from the other folders:

- A **design doc** looks at the whole milestone: the problem, the options, the trade-offs, and a diagram.
- An **ADR** in [decisions/](../decisions/) records one single choice that comes out of it.
- The **milestone file** in [milestones/](../milestones/) is the step list and its status.

## How to add one

1. Copy `_template.md`.
2. Name it like its milestone file: `m03-real-time-chat.md`.
3. Fill it in, then the owner reads it and asks questions before the code starts. Keep it short.

A design doc can change while the milestone is built. Write the change in it, with the reason.
