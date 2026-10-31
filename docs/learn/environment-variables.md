# Environment variables

## What it is

An environment variable is a named setting that a program receives when it starts, outside its code. For example, `PORT=8080`.

Analogy: the code is a recipe, and environment variables are what you write on the order slip ("extra spicy"). The same recipe works for every order.

## Why we use it

The same program must run on your laptop, in Docker Compose, in CI and later in Kubernetes, each with different settings. So settings come from the environment, and the code has sensible defaults. Secrets (passwords, keys) are passed the same way, so they are never written in the code.

## Tiny example

In Go:

```go
port := os.Getenv("PORT") // "" if PORT is not set
if port == "" {
	port = "8080"           // the default
}
```

In the terminal:

```sh
PORT=9000 some-program   # sets PORT for this one command only
```

The hello service reads two: `PORT` (default `8080`) and `DATABASE_URL` (no default, because it contains a password: the service stops with a clear message if it is missing).

In Docker Compose, they are set in `compose.yaml`, with secrets filled in from the git-ignored `.env` file.

## Alternatives

- **Config files (YAML, TOML):** good for many settings, but they need a file in each place the program runs.
- **Command-line flags (`--port=9000`):** clear, but awkward to pass in containers.

## Learn more

- https://12factor.net/config
