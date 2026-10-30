# Go

## What it is

Go is a programming language made at Google for servers and tools. You write `.go` files, and the Go compiler turns them into one program file (a "binary") that runs on its own, with nothing else to install.

## Why we use it

It is small enough to learn fully, fast, and very good at handling many connections at once. See [ADR 0003](../decisions/0003-use-go-for-services.md).

## Key ideas

- **Package:** a folder of `.go` files that belong together. Each file starts with `package name`. The package `main` is special: it builds a runnable program, which starts in the function `main()`.
- **Module:** a project with its own `go.mod` file. `go.mod` gives the module its name (its "module path") and lists the libraries it uses. Each service in `services/` is its own module.
- **Errors are values:** Go has no exceptions. A function that can fail returns an `error` as its last result, and the caller checks it right away with `if err != nil`.
- **Standard library:** Go comes with many ready-made packages, such as `net/http` for web servers and `log/slog` for logs. We prefer them over outside libraries.

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
