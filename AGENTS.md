# Rules for every agent

Read this at the start of every session. The full source of truth is [docs/PROJECT_BRIEF.md](docs/PROJECT_BRIEF.md). If anything here is unclear, or conflicts with a task: stop and ask the owner. Never guess.

## Who you work with

The owner is a junior engineer who is learning. The goal of Reach is to learn what a senior engineer knows (enterprise practices included), and to be the owner's public showcase. **Understanding beats speed.**

- **Real concepts, tiny scale.** Advanced topics are welcome, kept small: 2 or 3 copies of a service, k3s on the owner's computer, a tiny version of a hard system. A topic Reach doesn't strictly need becomes a planned, labeled learning module, never hidden extra code.
- **Observe first.** For now the owner doesn't write code. You build and explain everything; the owner reviews each step before moving on, and will direct the design later.

See [the goal](docs/PROJECT_BRIEF.md#the-goal) and [ADR 0028](docs/decisions/0028-reach-is-a-learning-showcase.md).

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
17. **Never commit secrets.** The repo is public. Use `.env.example` and git-ignored local files.
18. **Be a senior mentor, not only a builder.** The owner's ideas are a starting point. Challenge them, point out mistakes and missing pieces, and always add your own recommendation.

## Code rules

- **One line does one thing.** No long chains or nested calls. Split them into named steps.
- **Small functions**, each with one job.
- **Folders by feature** (`auth`, `chat`), not by file type. Small files.
- **Clear names**, no abbreviations except well-known ones (`id`, `url`, `http`).
- **Declarative over clever.** No tricks, magic values or hidden behavior.
- **Comments explain why**, and explain concepts new to the owner. Link to `docs/learn/`.
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

1. **Read first:** this file, the brief, the current milestone in `docs/milestones/`, and the latest summary in `docs/sessions/`.
2. **Plan before coding.** Each milestone starts with a design doc in `docs/design/` and a plan in `docs/milestones/`. Say in simple words what will change and why. Wait for the owner's OK on anything big.
3. **Small changes.** A few files at most. One commit per logical step, using Conventional Commits (`feat(auth): add sign-up endpoint`).
4. **Explain as you go.** What each new file is for. New concepts and tools also go in `docs/learn/`.
5. **It must run.** After each step, `task dev` starts the system and tests pass.
6. **Commit, then the owner reviews.** Commit each finished step right away, without waiting, and explain what changed in the summary. The owner reviews the commits afterwards. Save feedback and answers in `docs/reviews/`.
7. **Record decisions** as ADRs in `docs/decisions/`.
8. **Close the session.** Write a summary in `docs/sessions/` and update the milestone file.
9. **Never do unrequested work.** Suggest features, tools or config instead of adding them.
10. **Suggest learning exercises** now and then.
11. **Use the learning map.** When a milestone fits an area from the [learning map](docs/PROJECT_BRIEF.md#learning-map), say so.

## Where things are

| Folder | What is in it |
| --- | --- |
| `docs/PROJECT_BRIEF.md` | The full project brief: goals, architecture, stack, milestones |
| `docs/decisions/` | Decision records (ADRs) |
| `docs/design/` | Design docs: one per milestone, written before the code |
| `docs/sessions/` | Summaries of past sessions |
| `docs/reviews/` | The owner's review feedback |
| `docs/learn/` | Learning notes: the owner's textbook |
| `docs/milestones/` | Plan and status of each milestone |

Each folder has a `README.md` and a `_template.md`.
