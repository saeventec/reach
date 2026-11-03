# Code style

How individual lines and functions are written.

## Return early

- **What:** check for a problem, handle it, and return at once. The happy path continues at the left edge, without nesting.
- **Why:** each function reads top to bottom as a list of steps.
- **Example:**

```
user, err = findUser(id)
if err: return err

hash, err = loadHash(user)
if err: return err

return hash
```

## Small functions with one job

- **What:** a function does one thing, and its name says what. Bigger jobs call several small, named steps.
- **Why:** each step is easy to read, test and reuse.

## Explicit over magic

- **What:** prefer code you can read over tricks. Write data conversions field by field. Write startup order line by line. Avoid reflection-heavy or "auto" tools when plain code is short.
- **Why:** what you see is what runs.

## Named values, no magic numbers or strings

- **What:** give fixed values names: response codes, roles, context keys, cookie names, limits.
- **Why:** the name explains the value, and changing it happens in one place.
- **Example:**

```
const KeyUser   = "user"
const KeyAccess = "access_token"
const MaxPageSize = 2000
```

## Typed sets of choices

- **What:** when a field has a fixed set of values, define a type with named constants instead of using bare strings.
- **Why:** typos are caught, and the allowed values are listed in one place.
- **Example:**

```
type PostType: "news" | "review" | "article"
```

## Consistent naming patterns

- **What:** the same idea always gets the same name: `New…` to create, `Find` for one, `Filter` for many, `Save`, `Delete`, `Err…` for named errors, `…Request` and `…Response` for API data.
- **Why:** you can guess a name before you look it up.

## Comments explain why

- **What:** few comments, and only where the code can't say it: a decision, a limit, a trade-off, or a "this is temporary" note.
- **Why:** comments that repeat the code go stale. Comments that explain intent stay useful.
- **Example:**

```
// The expression can be a role ("admin") or role plus permission
// ("role=admin,perm=create"). The provider decides how to read it.
```

## Leave honest TODOs

- **What:** when something is knowingly temporary or unsafe, mark it with a `TODO` that says what's missing.
- **Why:** the gap is visible and searchable instead of forgotten.

## Formatting by tools and editor settings

- **What:** an editor config file fixes encoding, line endings, indentation, final newline and maximum line length for every file type. Code formatters run with default settings.
- **Why:** no style debates, and every file looks the same in every editor.
