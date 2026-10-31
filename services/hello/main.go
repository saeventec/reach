// The hello service: a small Go program that answers HTTP requests.
// It exists for milestone 1. See docs/decisions/0011-start-with-a-hello-service.md.
// Learn more: docs/learn/go.md and docs/learn/http.md
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	"github.com/saeventec/reach/services/hello/gen/hello/v1/hellov1connect"
)

// shutdownTimeout is how long open requests get to finish when the service
// is asked to stop. It is shorter than the 10 seconds Docker waits before
// it kills a container.
const shutdownTimeout = 5 * time.Second

// main is where every Go program starts. It only sets up logging and
// reports the error from run, so the real work stays easy to test.
func main() {
	logHandler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	err := run()
	if err != nil {
		slog.Error("the service stopped", "error", err)
		os.Exit(1)
	}
}

// run connects to the database, starts the HTTP server, and waits until
// the server fails or is asked to stop.
// Learn more: docs/learn/graceful-shutdown.md
func run() error {
	// stopSignal is "done" when the operating system asks us to stop:
	// Ctrl+C in a terminal (SIGINT), or "docker stop" (SIGTERM).
	stopSignal, stopListening := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopListening()

	databaseURL, err := readDatabaseURL()
	if err != nil {
		return err
	}

	pool, err := connectToDatabase(stopSignal, databaseURL)
	if err != nil {
		return err
	}
	// Close the database connections last, after the server has stopped.
	defer pool.Close()

	greetings := &postgresGreetingStore{pool: pool}

	port := readPort()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(greetings),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// The server runs in a goroutine (a function running at the same time as
	// this one), so run can wait for a stop signal meanwhile. The channel
	// carries the server's error back to us. It has room for one error, so
	// the goroutine can always send it and finish, even if nobody reads it.
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting the hello service", "address", address)
		err := server.ListenAndServe()
		serverErrors <- err
	}()

	// Wait for whichever comes first.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("serve http on %s: %w", address, err)
	case <-stopSignal.Done():
		slog.Info("stop signal received, shutting down")
	}

	return shutDown(server)
}

// shutDown stops accepting new requests, lets open requests finish, and
// gives up after shutdownTimeout.
func shutDown(server *http.Server) error {
	// A context with a deadline: Shutdown gives up when the time runs out.
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownContext)
	if err != nil {
		return fmt.Errorf("shut down the http server: %w", err)
	}

	slog.Info("the hello service stopped cleanly")
	return nil
}

// newRouter connects each URL path to the code that handles it.
func newRouter(greetings greetingStore) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", handleHealth)

	// The Connect server answers the API calls defined in proto/. Mount adds
	// one path per call to our router, like /hello.v1.GreetingService/GetGreeting.
	greetingService := &greetingServer{greetings: greetings}
	connectServer := connect.NewServer()
	hellov1connect.RegisterGreetingServiceHandler(connectServer, greetingService)
	connecthttp.Mount(router, connectServer)

	return router
}
