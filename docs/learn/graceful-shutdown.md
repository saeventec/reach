# Graceful shutdown

## What it is

Stopping a program politely: finish the work in progress, then exit. Like a shop that locks the door for new customers but still serves the ones inside.

The operating system asks a program to stop by sending a **signal**:

- **SIGINT** ("interrupt"): when you press Ctrl+C in a terminal.
- **SIGTERM** ("terminate"): when `docker stop` or Kubernetes stops a container. If the program hasn't exited after 10 seconds, Docker sends **SIGKILL**, which ends it at once, in the middle of whatever it was doing.

## Why we use it

Containers are stopped and started all the time: on every `task dev`, on updates, and when Kubernetes moves them. A graceful shutdown means no request is cut off halfway. It also stops faster: our service exits in a fraction of a second, instead of waiting 10 seconds to be killed.

## How the hello service does it

`run()` in `services/hello/main.go`:

1. `signal.NotifyContext` gives a **context** that becomes "done" when SIGINT or SIGTERM arrives. A context is Go's standard way to tell code "stop now" or "you have N seconds".
2. The server starts in a **goroutine**: a function that runs at the same time as the rest of the program. You start one with the word `go`.
3. A **channel** (`chan error`) carries the server's error back from the goroutine. A channel is a pipe between goroutines: one side sends with `channel <- value`, and the other side receives with `<-channel`.
4. `select` waits for whichever comes first: a server error, or the stop signal.
5. On a stop signal, `server.Shutdown` stops accepting new requests and waits for open ones, but gives up after 5 seconds (`context.WithTimeout`).

## Tiny example

```go
stopSignal, stopListening := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer stopListening() // "defer" runs this line when the function returns

select {
case err := <-serverErrors:   // the server failed to start or crashed
	return err
case <-stopSignal.Done():     // someone asked us to stop
	slog.Info("stop signal received, shutting down")
}
```

What you see in the logs after `docker compose stop hello`:

```
level=INFO msg="stop signal received, shutting down"
level=INFO msg="the hello service stopped cleanly"
hello-1 exited with code 0
```

Exit code 0 means "ended normally". Any other number means an error.

## Alternatives

- **No graceful shutdown:** the program is killed. Simple, but requests are cut off, and every stop takes 10 seconds.
- **Libraries that manage the program's lifecycle** (such as `oklog/run`): useful with many parts to stop, too much for one server.

## Learn more

- https://pkg.go.dev/net/http#Server.Shutdown
- https://go.dev/tour/concurrency/1 (goroutines, channels and select)
- https://docs.docker.com/reference/cli/docker/container/stop/
