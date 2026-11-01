package main

// The WebSocket endpoint, GET /realtime. The browser opens one WebSocket
// after logging in, and the server pushes events on it.
// Learn more: docs/learn/websocket.md

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// userIDHeader carries the logged-in user's id. Only the gateway sets it,
// after asking auth (ADR 0029).
const userIDHeader = "X-User-Id"

// writeTimeout is how long writing one event may take. A browser that
// doesn't take it in time is disconnected.
const writeTimeout = 10 * time.Second

// SocketHandler accepts WebSockets and connects them to the hub.
type SocketHandler struct {
	hub *Hub
	// stopping is done when the service is asked to stop. Then every
	// WebSocket is closed with "going away", so browsers reconnect.
	stopping context.Context
	// open counts the open WebSockets, so the service can wait for them to
	// close before it exits. See main.go.
	open sync.WaitGroup
}

// NewSocketHandler creates the handler.
func NewSocketHandler(hub *Hub, stopping context.Context) *SocketHandler {
	return &SocketHandler{hub: hub, stopping: stopping}
}

// ServeHTTP answers GET /realtime. It runs until the WebSocket closes.
func (handler *SocketHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Counted before anything else: see Wait.
	handler.open.Add(1)
	defer handler.open.Done()

	userID := request.Header.Get(userIDHeader)
	if userID == "" {
		// The gateway always sets the header for a logged-in user. Without
		// it, the request didn't come through the gateway's session check.
		http.Error(writer, "you are not logged in", http.StatusUnauthorized)
		return
	}

	// Accept turns the HTTP request into a WebSocket ("the upgrade"). With
	// no options it refuses a page from another website (its Origin header
	// doesn't match the address it connects to) and writes the error
	// answer itself. Learn more: docs/learn/websocket.md#the-origin-check
	connection, err := websocket.Accept(writer, request, nil)
	if err != nil {
		slog.Warn("could not accept the websocket", "error", err)
		return
	}

	userClient := newClient(userID)
	handler.hub.add(userClient)
	defer handler.hub.remove(userClient)

	handler.sendEvents(connection, userClient)
}

// Wait waits until every WebSocket has closed. Call it after the HTTP
// server has shut down: then no new WebSocket can start, because Shutdown
// waits for requests that haven't become WebSockets yet, and each request
// is counted before it becomes one.
func (handler *SocketHandler) Wait() {
	handler.open.Wait()
}

// sendEvents writes the client's events to the WebSocket until it closes:
// the browser leaves, the hub drops it, or the service stops.
func (handler *SocketHandler) sendEvents(connection *websocket.Conn, userClient *client) {
	// The browser sends nothing, but something must read, to notice when
	// it closes and to answer the WebSocket's own ping and close messages.
	// CloseRead does that in the background. Its context is done when the
	// WebSocket closes. (Not "stopping" as its parent: when its context
	// ends, the library cuts the connection at once, with no goodbye.)
	closed := connection.CloseRead(context.Background())

	for {
		select {
		case <-closed.Done():
			// The browser closed it (a tab was closed).
			return
		case <-handler.stopping.Done():
			// Say so, so the browser reconnects (maybe to another copy of
			// realtime). Close waits up to 10 seconds for the browser's
			// answer; browsers answer at once. The error is ignored on
			// purpose: the WebSocket is gone either way.
			_ = connection.Close(websocket.StatusGoingAway, "the server is restarting")
			return
		case frame, isOpen := <-userClient.frames:
			if !isOpen {
				// The hub removed this client: its buffer was full.
				_ = connection.Close(websocket.StatusTryAgainLater, "too many events waiting; reconnect")
				return
			}
			err := writeFrame(connection, frame)
			if err != nil {
				// A failed write has already closed the WebSocket.
				slog.Info("could not write to the websocket", "error", err)
				return
			}
		}
	}
}

// writeFrame writes one event as a text message, with a time limit.
func writeFrame(connection *websocket.Conn, frame []byte) error {
	// Not "stopping" as the parent: a write in progress may finish while
	// the service stops.
	writeContext, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	return connection.Write(writeContext, websocket.MessageText, frame)
}
