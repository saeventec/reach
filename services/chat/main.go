// The chat service: conversations, messages, and delivered and read status
// (milestone 3). Its code is organized in the hexagonal style ("ports and
// adapters", ADR 0030): main.go only connects the parts.
// Learn more: docs/design/m03-real-time-chat.md
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// Shared code, from packages/go. Learn more: docs/learn/go.md#sharing-code-between-modules
	"github.com/saeventec/reach/packages/go/config"
	"github.com/saeventec/reach/packages/go/database"
	"github.com/saeventec/reach/packages/go/health"
	"github.com/saeventec/reach/packages/go/httpserver"
)

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

// run connects to the database, then runs the HTTP server until it fails or
// is asked to stop.
// Learn more: docs/learn/graceful-shutdown.md
func run() error {
	// stopSignal is "done" when the operating system asks us to stop:
	// Ctrl+C in a terminal (SIGINT), or "docker stop" (SIGTERM).
	stopSignal, stopListening := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopListening()

	databaseURL, err := config.DatabaseURL()
	if err != nil {
		return err
	}

	// Nothing reads the database yet (the tables come in the next step).
	// Connecting now still finds a wrong address or password at startup.
	pool, err := database.Connect(stopSignal, databaseURL)
	if err != nil {
		return err
	}
	// Close the database connections last, after the server has stopped.
	defer pool.Close()

	port := config.Port()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	return httpserver.Run(stopSignal, "chat", server)
}

// newRouter connects each URL path to the code that handles it.
func newRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", health.Handle)
	return router
}
