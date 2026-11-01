# Go

## What it is

Go is a programming language made at Google for servers and tools. You write `.go` files, and the Go compiler turns them into one program file (a "binary") that runs on its own, with nothing else to install.

## Why we use it

It is small enough to learn fully, fast, and very good at handling many connections at once. See [ADR 0003](../decisions/0003-use-go-for-services.md).

## Key ideas

- **Package:** a folder of `.go` files that belong together. Each file starts with `package name`. The package `main` is special: it builds a runnable program, which starts in the function `main()`.
- **Module:** a project with its own `go.mod` file. `go.mod` gives the module its name (its "module path") and lists the libraries it uses. Each service in `services/` is its own module.
- **Errors are values:** Go has no exceptions. A function that can fail returns an `error` as its last result, and the caller checks it right away with `if err != nil`.
- **Struct and methods:** a `struct` groups named fields (`type greetingServer struct { greetings greetingStore }`). A method is a function attached to a type: `func (server *greetingServer) GetGreeting(...)`.
- **Interface:** a list of methods. Any type that has those methods fits the interface automatically, with no `implements` keyword. This is how a test can pass in a fake instead of the real database. See [testing.md](testing.md#fakes).
- **Pointer** (`*` and `&`): `&value` is the address of a value, and `*Type` is "an address of a Type". A function that gets a pointer can change the original value, and big values aren't copied.
- **Standard library:** Go comes with many ready-made packages, such as `net/http` for web servers and `log/slog` for logs. We prefer them over outside libraries.

## Sharing code between modules

Every service needs the same small things: read its settings, answer `/healthz`, connect to Postgres, stop cleanly. Instead of a copy in each service, they live once in the module `packages/go` ([ADR 0035](../decisions/0035-share-go-code-in-packages-go.md)).

- **Import path:** a package is imported by its module path plus its folder: `github.com/saeventec/reach/packages/go/config`. In the code, you then use the last part as the name: `config.Port()`.
- **Exported names:** only names that start with a **capital letter** can be used from another package. `config.Port` works; a lowercase `config.defaultPort` doesn't. Inside `package main` we kept everything lowercase, because nothing else imports it.
- **`replace`:** Go normally downloads a module from the internet, using its path. Our shared module isn't published, so each service's `go.mod` says "use this folder instead":

```
require github.com/saeventec/reach/packages/go v0.0.0   // v0.0.0: there are no versions, only the folder

replace github.com/saeventec/reach/packages/go => ../../packages/go
```

A change in `packages/go` reaches every service on its next build, so the code there stays small and tested.

**Alternatives:** a `go.work` file (Go "workspaces": one file lists all modules, nice for editors, awkward in Docker builds), publishing the module with versions (the normal way between separate repos), or copying (fine for two copies, not four).

## Tiny example

`services/hello/go.mod`:

```
module github.com/saeventec/reach/services/hello   // the module's name

go 1.27.3                                           // the Go version it needs
```

Handling an error:

```go
err := server.ListenAndServe()
if err != nil {
	// fmt.Errorf adds context. %w keeps the original error inside it.
	return fmt.Errorf("serve http on %s: %w", address, err)
}
```

## Commands

```sh
go run .         # compile and run the program in this folder
go test ./...    # run all tests in this folder and its subfolders
go vet ./...     # find common mistakes
gofmt -l .       # list files that are not formatted the standard way
go mod init <path>  # create a new module (writes go.mod)
```

## Alternatives

- **Node.js / TypeScript:** one language for front and back, but slower for heavy servers.
- **Rust:** very fast and safe, but much harder to learn.
- **Kotlin or C#:** good languages, but they need a large runtime (JVM or .NET).

## Learn more

- https://go.dev/tour (interactive tour, start here)
- https://go.dev/doc/effective_go
