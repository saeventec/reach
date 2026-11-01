// The realtime service: it keeps the browsers' WebSockets open, receives
// chat's events from NATS, and pushes each event to the WebSockets of the
// people it concerns (milestone 3, ADR 0033). It has no database: who is
// connected lives only in memory, and browsers reconnect after a restart.
// A small service, so all files are in one folder, like auth.
// Learn more: docs/learn/websocket.md
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

	"github.com/nats-io/nats.go"

	// Shared code, from packages/go. Learn more: docs/learn/go.md#sharing-code-between-modules
	"github.com/saeventec/reach/packages/go/chatevents"
	"github.com/saeventec/reach/packages/go/config"
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

// run connects the parts, then runs the HTTP server until it fails or is
// asked to stop. Learn more: docs/learn/graceful-shutdown.md
func run() error {
	// stopSignal is "done" when the operating system asks us to stop:
	// Ctrl+C in a terminal (SIGINT), or "docker stop" (SIGTERM).
	stopSignal, stopListening := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopListening()

	natsURL, err := config.NATSURL()
	if err != nil {
		return err
	}

	// nats.Connect reconnects on its own if the connection drops later.
	natsConnection, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("connect to nats at %s: %w", natsURL, err)
	}
	// Closing also ends the subscription below.
	defer natsConnection.Close()

	hub := NewHub()
	events := NewEventHandler(hub)
	sockets := NewSocketHandler(hub, stopSignal)

	// A plain ("core") NATS subscription: realtime gets each event that is
	// published while it runs. Each copy of realtime gets every event, and
	// delivers it to its own WebSockets. Events published while it was
	// down are not needed: browsers catch up through chat's API.
	_, err = natsConnection.Subscribe(chatevents.All, events.HandleMessage)
	if err != nil {
		return fmt.Errorf("subscribe to %s: %w", chatevents.All, err)
	}

	port := config.Port()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(sockets),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	err = httpserver.Run(stopSignal, "realtime", server)
	if err != nil {
		// The server failed: no clean goodbye, the program ends now.
		return err
	}

	// Shutdown doesn't wait for WebSockets: once a request has become a
	// WebSocket, the HTTP server no longer tracks it. They close themselves
	// when stopSignal is done (see socket.go); wait until they have.
	sockets.Wait()
	return nil
}

// newRouter connects each URL path to the code that handles it.
func newRouter(sockets *SocketHandler) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", health.Handle)
	router.Handle("GET /realtime", sockets)
	return router
}
