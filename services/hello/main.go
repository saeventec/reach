// The hello service: a small Go program that answers HTTP requests.
// It exists for milestone 1. See docs/decisions/0011-start-with-a-hello-service.md.
// Learn more: docs/learn/go.md and docs/learn/http.md
package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"
)

// defaultPort is used when the PORT environment variable is empty.
const defaultPort = "8080"

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

// run starts the HTTP server and blocks until the server stops.
func run() error {
	port := readPort()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("starting the hello service", "address", address)

	err := server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("serve http on %s: %w", address, err)
	}
	return nil
}

// readPort returns the port to listen on, from the PORT environment variable.
// Configuration comes from the environment, so the same program runs
// anywhere. Learn more: docs/learn/environment-variables.md
func readPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return defaultPort
	}
	return port
}

// newRouter connects each URL path to the function that handles it.
func newRouter() *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", handleHealth)
	return router
}
