# Rules for every agent

Reach is a chat app built as a **learning project and public showcase** for its owner, a junior engineer ([ADR 0028](docs/decisions/0028-reach-is-a-learning-showcase.md)). This file holds all the rules. If a rule is unclear or conflicts with a task: stop and ask the owner. Never guess.

## Reading docs

This file is the only thing to read at the start. Everything else is read **on demand**:

- Pick the file from [Where things are](#where-things-are) or the folder's `README.md` index, then read only the section you need.
- Continuing earlier work: read the status in [docs/milestones/README.md](docs/milestones/README.md) and the latest session summary.
- Read a whole long file, or search all docs, only when the task needs it or the owner asks.

## Who you work with

- **Understanding beats speed.** The owner reviews and learns every step. A slower, clearer solution wins.
- **Real concepts, tiny scale.** Advanced topics are welcome, kept small (2 or 3 copies of a service, k3s on one computer). A topic Reach doesn't need becomes a planned, labeled learning module, never hidden extra code.
- **Observe first.** For now the owner doesn't write code. You build and explain; the owner reviews, and will direct the design later.

## Golden rules

1. **Assume no background knowledge.** Never assume the owner knows a tool, concept, command or term.
2. **Never leave anything unexplained.** Every new tool, concept, library, command, config option and pattern gets a short, simple explanation the first time it appears.
3. **Explain in simple terms.** Plain English, short sentences, small examples. Define any jargon.
4. **Always show alternatives.** For each tech choice: what it is, why we chose it, 2 to 4 alternatives (one line each), and a tiny example.
5. **The owner must understand every single line.** If a line can't be explained simply, rewrite it.
6. **Simple, clean, small and organized.** No cleverness, no premature optimization, no "just in case" code.
7. **Current stable versions and best practices.** Check official, up-to-date docs before writing code or config. Never use deprecated APIs.
8. **Everything as code.** No manual setup. If it isn't in the repo, it doesn't exist.
9. **Self-hosted.** No cloud platforms. Allowed: GitHub, app stores, Apple APNs, Google FCM.
10. **Runs on one computer** with Docker Compose and one command, on Mac and Windows (WSL2).
11. **One repository** (monorepo) for code, infrastructure, docs and decisions.
12. **Use defaults.** Only add configuration that is truly needed, and explain each setting.
13. **Document everything, in git**, in `docs/`.
14. **Small steps.** One small, reviewable change at a time. Each step ends with something that runs.
15. **Ask when unsure.** Understand the owner's real goal, not only the words.
16. **Stay flexible.** Design so a part can be swapped without rewriting everything.
17. **Never commit secrets or private personal data.** The repo is public. Use `.env.example` and git-ignored local files.
18. **Be a senior mentor, not only a builder.** The owner's ideas are a starting point. Challenge them, point out mistakes and missing pieces, and always add your own recommendation.
19. **Readable beats short.** Keep every block of code small, with one clear job. `main` and handlers only list the steps; each step is its own named function. If a section needs a heading comment, make it a function.
20. **Token-efficient docs.** Write for fast lookup: one topic per file, clear headings, facts first, lists and tables over prose. Explain once, link everywhere else. Keep each folder's `README.md` index current.

## Code rules

- **One line does one thing.** No long chains or nested calls. Split them into named steps.
- **Small functions**, each with one job, about 30 lines at most. Split a file when it mixes jobs.
- **Folders by feature** (`auth`, `chat`), not by file type. Small files.
- **Clear names**, no abbreviations except well-known ones (`id`, `url`, `http`).
- **Declarative over clever.** No tricks, magic values or hidden behavior.
- **Comments explain why**, in a line or two. Concepts are explained in `docs/learn/` and linked, not repeated in the code.
- **Explicit error handling.** Clear messages. Never ignore an error silently.
- **Few dependencies.** Prefer the standard library. Explain every library added.
- **Automatic formatting and linting** with default settings (`gofmt`, golangci-lint, Oxlint, Prettier, buf).
- **Tests from day one**, next to the code, readable like examples.

```go
// Bad: several things in one line
user, err := repo.Find(strings.ToLower(strings.TrimSpace(req.Email)))

// Good: one step per line, clear names
email := strings.TrimSpace(request.Email)
normalizedEmail := strings.ToLower(email)
user, err := userRepository.FindByEmail(normalizedEmail)
```

## Workflow

1. **Read on demand:** see [Reading docs](#reading-docs).
2. **Plan before coding.** Each milestone starts with a design doc in `docs/design/` and a plan in `docs/milestones/`. Say in simple words what will change and why. Wait for the owner's OK on anything big.
3. **Small changes.** A few files at most. One commit per logical step, using Conventional Commits (`feat(auth): add sign-up endpoint`).
4. **Explain as you go.** What each new file is for. New concepts and tools also go in `docs/learn/`.
5. **It must run.** After each step, `task dev` starts the system and tests pass.
6. **Commit, then the owner reviews.** Commit each finished step right away and explain what changed. The owner reviews afterwards. Save feedback and answers in `docs/reviews/`.
7. **Record decisions** as ADRs in `docs/decisions/`, and update the brief if it changes the plan.
8. **Close the session.** Write a short summary in `docs/sessions/`, and update the milestone file and the folder indexes.
9. **Never do unrequested work.** Suggest features, tools or config instead of adding them.
10. **Suggest learning exercises** now and then.
11. **Use the learning map.** When a milestone fits an area from the [learning map](docs/PROJECT_BRIEF.md#learning-map), say so.

## Where things are

| Path | What is in it | Read it when |
| --- | --- | --- |
| [docs/PROJECT_BRIEF.md](docs/PROJECT_BRIEF.md) | Goals, architecture, stack, milestones, future ideas | Planning, or a question about scope or architecture (one section) |
| [docs/milestones/](docs/milestones/README.md) | Plan and status of each milestone | Starting or continuing a milestone |
| [docs/sessions/](docs/sessions/README.md) | Short summaries of past sessions | Continuing earlier work (the latest one) |
| [docs/design/](docs/design/README.md) | One design doc per milestone, written before the code | Working on that milestone |
| [docs/decisions/](docs/decisions/README.md) | Decision records (ADRs) | Before changing something an ADR decided |
| [docs/learn/](docs/learn/README.md) | Learning notes: the owner's textbook | Introducing or linking a concept |
| [docs/reviews/](docs/reviews/README.md) | The owner's review feedback | Answering or applying a review |
| [docs/style/](docs/style/README.md) | The owner's habits and patterns | Designing new code or a refactor |

Each folder has a `README.md` (its index) and a `_template.md`.
