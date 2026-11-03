# Ideas for Reach

Where Reach's code differs from the habits in this folder today, and what we might change in a future step. **Nothing here is decided.** Each idea needs the owner's OK, and a big one needs a design doc and an ADR first.

| # | Idea | Reach today | Habit | Size |
| --- | --- | --- | --- | --- |
| 1 | One shape for every service | `auth`, `hello` and `realtime` are flat folders. `chat` has layered sub-packages | [One shape, repeated everywhere](organization.md#one-shape-repeated-everywhere) | Big (a refactor) |
| 2 | Features register their own routes | Each `main.go` builds and registers everything in `newRouter` | [Each feature registers itself](design.md#each-feature-registers-itself) | Medium |
| 3 | One settings struct per service | Settings are read by separate functions, some shared, some per service | [One settings struct, read once](configuration.md#one-settings-struct-read-once) | Small |
| 4 | One shared error vocabulary | `chat` maps errors in one place. `auth` builds error responses inline in each handler | [Send responses through one helper](errors-and-responses.md#send-responses-through-one-helper) | Medium |
| 5 | Always use constructors | `chat` uses `New…` functions. `auth` builds structs directly | [Constructors for everything](design.md#constructors-for-everything) | Small |
| 6 | Shorter comments in code | Long concept explanations sit inline in the code | **Decided for new code** ([ADR 0036](../decisions/0036-readable-code-and-token-efficient-docs.md)). Existing comments: move concepts to `docs/learn/`, keep a link | Medium |
| 7 | Short method-owner names | Methods use full names (`service *Service`) | The language's own convention is one or two letters. Would need an exception to "no abbreviations" | Small |
| 8 | Small `main` files | Each `main.go` reads settings, connects, builds parts, starts workers and routes in one long function | **Decided for new code** (golden rule 19). Existing files: split into named steps | Medium |
| 9 | Tighter docs | Some docs are long or repeat explanations | **Decided for new docs** (golden rule 20). Existing docs: shorten, link instead of repeating | Medium |

## Habits Reach already follows

- Features grouped in folders, building blocks shared separately (`packages/go`).
- Small interfaces, with fakes in tests.
- Wiring by hand in one place (`main.go`).
- Settings from the environment, `.env.example`, no defaults for secrets.
- Startup fails fast when the database doesn't answer.
- Internal errors are logged, and clients get a general message.
- Editor config, command runner with a list of commands, compose, two-stage images, CI.

## Where Reach intentionally differs

These differences are on purpose, and we keep them unless a new decision says otherwise:

- **Full names** instead of short ones: easier to read for someone new.
- **Every error handled** explicitly, never ignored.
- **SQL with generated code** instead of an ORM ([ADR 0012](../decisions/0012-use-pgx-for-database-access.md)).
- **Sessions stored on the server** instead of self-contained tokens, so a device can be logged out at once ([ADR 0022](../decisions/0022-keep-sessions-in-httponly-cookies.md)).
- **Standard-library tests** instead of an assertion library: fewer dependencies.
