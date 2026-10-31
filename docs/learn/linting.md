# Linting and formatting

## What it is

- A **linter** reads your code without running it, and points out likely bugs and risky patterns. For example: an error that is never checked, or a variable that is never used.
- A **formatter** rewrites code into one standard layout (spaces, line breaks, quotes). Nobody argues about style, and diffs only show real changes.

Analogy: the linter is a proofreader that flags mistakes; the formatter is a typesetter that makes every page look the same.

## Why we use it

Mistakes are found in seconds, before review and before CI. We use each tool's default rules (Golden rule 12).

## Commands

```sh
task lint     # check everything; changes nothing; fails if something is wrong
task format   # rewrite files into the standard format
```

Each part also has its own task, like `task lint:go`. Run `task` to see them all.

## Go: golangci-lint

golangci-lint runs many Go linters at once. We use its default set:

| Linter | Finds |
| --- | --- |
| `errcheck` | errors that are never checked |
| `govet` | suspicious code (the same checks as `go vet`) |
| `ineffassign` | values assigned but never used |
| `staticcheck` | bugs, simplifications, deprecated APIs |
| `unused` | functions, types and variables nobody uses |

We also turned on its `gofmt` formatter in `.golangci.yml`, so `task lint` reports files that aren't in Go's standard format, and `task format` fixes them. Generated files (`DO NOT EDIT`) are skipped automatically.

Example output:

```
hello/broken.go:6:11: Error return value of `os.Remove` is not checked (errcheck)
	os.Remove("x")
	         ^
```

Read it as: file, line, column, the problem, and the linter's name in brackets.

## Alternatives

- **`go vet` + `gofmt` alone:** built into Go, fewer checks.
- **staticcheck alone:** very good, but one linter. golangci-lint already includes it.

## Learn more

- https://golangci-lint.run/docs/
