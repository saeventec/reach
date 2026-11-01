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
	"sync"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	// Shared code, from packages/go. Learn more: docs/learn/go.md#sharing-code-between-modules
	"github.com/saeventec/reach/packages/go/config"
	"github.com/saeventec/reach/packages/go/database"
	"github.com/saeventec/reach/packages/go/gen/chat/v1/chatv1connect"
	"github.com/saeventec/reach/packages/go/health"
	"github.com/saeventec/reach/packages/go/httpserver"

	"github.com/saeventec/reach/services/chat/api"
	"github.com/saeventec/reach/services/chat/authclient"
	"github.com/saeventec/reach/services/chat/conversation"
	"github.com/saeventec/reach/services/chat/natspublisher"
	"github.com/saeventec/reach/services/chat/outbox"
	"github.com/saeventec/reach/services/chat/postgres"
)

// maximumRequestBytes is the largest API message the service reads. The
// biggest real one is a message of 4000 characters: a character takes up
// to 4 bytes, so up to 16 000 bytes, plus the rest of the request. Bigger
// messages get a RESOURCE_EXHAUSTED error.
const maximumRequestBytes = 32 * 1024

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

// run connects the parts, then runs the outbox relay and the HTTP server
// until the server fails or the service is asked to stop.
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

	authURL, err := readAuthURL()
	if err != nil {
		return err
	}

	natsURL, err := readNATSURL()
	if err != nil {
		return err
	}

	pool, err := database.Connect(stopSignal, databaseURL)
	if err != nil {
		return err
	}
	// Close the database connections last, after the server has stopped.
	defer pool.Close()

	publisher, err := natspublisher.Connect(stopSignal, natsURL)
	if err != nil {
		return err
	}
	defer publisher.Close()

	// The store sends on this channel after saving an event, to wake the
	// relay at once. Room for one value: see Store.wakeRelay.
	eventSaved := make(chan struct{}, 1)

	// The hexagon: the core gets its two adapters (the database and the
	// auth service), and the API adapter gets the core. The relay reads
	// the events from the database and publishes them to NATS.
	// Learn more: docs/learn/hexagonal-architecture.md
	store := postgres.NewStore(pool, eventSaved)
	directory := authclient.NewDirectory(authURL)
	conversations := conversation.NewService(store, directory)
	relay := outbox.NewRelay(store, publisher, eventSaved)

	port := config.Port()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(conversations),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	// The relay runs in its own goroutine, next to the server. Its context
	// is stopped when the service stops, and also if the server fails.
	relayContext, stopRelay := context.WithCancel(stopSignal)
	var relayRunning sync.WaitGroup
	relayRunning.Go(func() {
		relay.Run(relayContext)
	})

	err = httpserver.Run(stopSignal, "chat", server)

	// Wait for the relay to finish its work before the deferred lines
	// close NATS and the database.
	stopRelay()
	relayRunning.Wait()
	return err
}

// newRouter connects each URL path to the code that handles it.
func newRouter(conversations *conversation.Service) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", health.Handle)

	// The Connect server answers the API calls defined in proto/chat/v1,
	// like /chat.v1.ConversationService/ListConversations and
	// /chat.v1.MessageService/SendMessage.
	conversationServer := api.NewConversationServer(conversations)
	connectServer := connect.NewServer()
	chatv1connect.RegisterConversationServiceHandler(connectServer, conversationServer)
	messageServer := api.NewMessageServer(conversations)
	chatv1connect.RegisterMessageServiceHandler(connectServer, messageServer)
	connecthttp.Mount(router, connectServer, connecthttp.WithReadMaxBytes(maximumRequestBytes))

	return router
}
