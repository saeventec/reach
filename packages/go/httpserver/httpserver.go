// Package httpserver runs a service's HTTP server and stops it cleanly.
// Learn more: docs/learn/graceful-shutdown.md
package httpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// shutdownTimeout is how long open requests get to finish when the service
// is asked to stop. It is shorter than the 10 seconds Docker waits before
// it kills a container.
const shutdownTimeout = 5 * time.Second

// Run starts the server and waits until it fails or stopSignal is done.
// Then it shuts the server down cleanly. serviceName is only for the logs.
func Run(stopSignal context.Context, serviceName string, server *http.Server) error {
	// The server runs in a goroutine (a function running at the same time as
	// this one), so Run can wait for the stop signal meanwhile. The channel
	// carries the server's error back to us. It has room for one error, so
	// the goroutine can always send it and finish, even if nobody reads it.
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("starting the service", "service", serviceName, "address", server.Addr)
		err := server.ListenAndServe()
		serverErrors <- err
	}()

	// Wait for whichever comes first.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("serve http on %s: %w", server.Addr, err)
	case <-stopSignal.Done():
		slog.Info("stop signal received, shutting down", "service", serviceName)
	}

	return shutDown(serviceName, server)
}

// shutDown stops accepting new requests, lets open requests finish, and
// gives up after shutdownTimeout.
func shutDown(serviceName string, server *http.Server) error {
	// A context with a deadline: Shutdown gives up when the time runs out.
	shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	err := server.Shutdown(shutdownContext)
	if err != nil {
		return fmt.Errorf("shut down the http server: %w", err)
	}

	slog.Info("the service stopped cleanly", "service", serviceName)
	return nil
}
