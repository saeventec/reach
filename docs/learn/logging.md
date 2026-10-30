# Logging

## What it is

A log is a list of short messages a program writes while it runs: "started", "request failed", and so on. Logs are how you find out what happened when something goes wrong.

**Structured logs** write each message with named fields (`key=value`), not just a sentence. Tools can then search and filter them, for example "all errors from the hello service".

## Why we use it

Every service logs to standard output (the terminal, also called "stdout"). Docker collects stdout automatically, and later the observability milestone sends it to a log system. The program itself never deals with log files.

## Tiny example

Go's built-in `log/slog` package:

```go
logHandler := slog.NewTextHandler(os.Stdout, nil) // write key=value lines to stdout
logger := slog.New(logHandler)
slog.SetDefault(logger)                          // slog.Info, slog.Error... now use it

slog.Info("starting the hello service", "address", ":8080")
```

Output:

```
time=2026-10-31T02:53:57.358+03:30 level=INFO msg="starting the hello service" address=:8080
```

**Levels** show how important a message is: `DEBUG`, `INFO`, `WARN`, `ERROR`.

We use the text format because it is easy to read. Switching to JSON later is one line: `slog.NewJSONHandler`.

## Alternatives

- **zap or zerolog:** popular, very fast logging libraries, but `slog` is built in and enough.
- **`log` (the older built-in package):** plain sentences, no fields.

## Learn more

- https://go.dev/blog/slog
- https://pkg.go.dev/log/slog
