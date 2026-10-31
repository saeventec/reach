# 0017. Use buf with local generators, and commit the generated code

- **Status:** Accepted
- **Date:** 2026-10-31

## Context

[ADR 0005](0005-protobuf-and-connect-for-the-api.md) chose Protobuf and Connect, and left the code generation tool open. We need to check `.proto` files and generate Go (and soon TypeScript) code from them.

## Decision

Use buf. Run the generators locally (`local:` in `buf.gen.yaml`), installed by mise. Commit the generated code to git. Regenerate it with `task generate`.

## Why

- buf is the standard Protobuf tool today: lint, breaking-change checks and generation in one, with simple YAML config.
- Local generators keep us self-hosted (Golden rule 9). buf's "remote plugins" would run on buf's cloud service.
- Committed code: Docker builds and new clones work without buf, and you can read and review exactly what was generated.

## Alternatives considered

- **protoc:** the original compiler. It works, but has long command lines, no lint and no breaking-change checks.
- **buf remote plugins:** no generators to install, but needs buf's cloud service.
- **Generate during the Docker build, don't commit:** a cleaner git history, but every build needs buf, and the generated code is never reviewed.

## Consequences

After changing a `.proto` file, always run `task generate` and commit the result. CI (step 8) will check that the generated code is up to date. The generated Go code lives in `services/hello/gen` for now. When a second service needs it, it moves to shared code in `packages/go`.
