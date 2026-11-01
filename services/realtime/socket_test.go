package main

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// testServer runs the router on a real local port, like the service does.
// stop is like the service's stop signal.
type testServer struct {
	url     string
	hub     *Hub
	sockets *SocketHandler
	stop    context.CancelFunc
}

func startTestServer(t *testing.T) testServer {
	t.Helper()
	stopping, stop := context.WithCancel(context.Background())
	hub := NewHub()
	sockets := NewSocketHandler(hub, stopping)

	// httptest.NewServer starts a real HTTP server on a free port.
	server := httptest.NewServer(newRouter(sockets))
	t.Cleanup(func() {
		stop()
		sockets.Wait()
		server.Close()
	})

	// WebSocket addresses start with ws:// instead of http://.
	url := strings.Replace(server.URL, "http://", "ws://", 1) + "/realtime"
	return testServer{url: url, hub: hub, sockets: sockets, stop: stop}
}

// dial opens a WebSocket with these headers, as the gateway would send.
func dial(t *testing.T, url string, header http.Header) (*websocket.Conn, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	options := &websocket.DialOptions{HTTPHeader: header}
	return websocket.Dial(ctx, url, options)
}

// dialAs opens a WebSocket as this user, and waits until the hub has it.
func dialAs(t *testing.T, server testServer, userID string) *websocket.Conn {
	t.Helper()
	before := server.hub.connectedCount()
	header := http.Header{userIDHeader: []string{userID}}

	connection, _, err := dial(t, server.url, header)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = connection.CloseNow() })

	// The server adds the client just after the handshake, so wait a
	// moment for it.
	waitUntil(t, func() bool { return server.hub.connectedCount() > before })
	return connection
}

// waitUntil checks the condition every few milliseconds, for up to two
// seconds, and fails the test if it never becomes true.
func waitUntil(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the condition never became true")
}

// readText reads one message from the WebSocket.
func readText(t *testing.T, connection *websocket.Conn) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()
	_, data, err := connection.Read(ctx)
	return string(data), err
}

func TestSocket(t *testing.T) {
	t.Run("pushes the user's events", func(t *testing.T) {
		server := startTestServer(t)
		bob := dialAs(t, server, "bob")

		server.hub.Deliver([]string{"bob"}, []byte(`{"hello":"bob"}`))

		got, err := readText(t, bob)
		if err != nil {
			t.Fatalf("read: %v", err)
		}
		if got != `{"hello":"bob"}` {
			t.Errorf("got %s, want {\"hello\":\"bob\"}", got)
		}
	})

	t.Run("refuses a request without the gateway's user header", func(t *testing.T) {
		server := startTestServer(t)

		_, response, err := dial(t, server.url, nil)

		if err == nil {
			t.Fatal("connected, want a refusal")
		}
		if response == nil || response.StatusCode != http.StatusUnauthorized {
			t.Errorf("got %v, want status 401", response)
		}
	})

	t.Run("refuses a page from another website", func(t *testing.T) {
		server := startTestServer(t)
		// Another website's page, open in the user's browser, tries to
		// connect with the user's cookie: its Origin gives it away.
		header := http.Header{
			userIDHeader: []string{"bob"},
			"Origin":     []string{"https://evil.example"},
		}

		_, response, err := dial(t, server.url, header)

		if err == nil {
			t.Fatal("connected, want a refusal")
		}
		if response == nil || response.StatusCode != http.StatusForbidden {
			t.Errorf("got %v, want status 403", response)
		}
	})

	t.Run("leaves the hub when the browser closes", func(t *testing.T) {
		server := startTestServer(t)
		bob := dialAs(t, server, "bob")

		_ = bob.Close(websocket.StatusNormalClosure, "tab closed")

		waitUntil(t, func() bool { return server.hub.connectedCount() == 0 })
	})

	t.Run("says going away when the service stops", func(t *testing.T) {
		server := startTestServer(t)
		bob := dialAs(t, server, "bob")

		server.stop()

		_, err := readText(t, bob)
		status := websocket.CloseStatus(err)
		if status != websocket.StatusGoingAway {
			t.Errorf("got close status %v (error %v), want %v", status, err, websocket.StatusGoingAway)
		}
		// Wait returns: every WebSocket is closed.
		server.sockets.Wait()
	})

	t.Run("closes a browser that falls too far behind", func(t *testing.T) {
		server := startTestServer(t)
		bob := dialAs(t, server, "bob")

		// Fill bob's buffer faster than anyone reads it. The server's write
		// loop takes some frames out meanwhile, so send plenty.
		for range clientBufferSize * 100 {
			server.hub.Deliver([]string{"bob"}, []byte("hello"))
		}

		// Read until the close arrives. The frames before it don't matter.
		var err error
		for err == nil {
			_, err = readText(t, bob)
		}
		var closeError websocket.CloseError
		if !errors.As(err, &closeError) || closeError.Code != websocket.StatusTryAgainLater {
			t.Errorf("got %v, want close status %v", err, websocket.StatusTryAgainLater)
		}
	})
}
