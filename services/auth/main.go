// The auth service: accounts, login, sessions and devices (milestone 2).
// Its base is copied from the hello service.
// Learn more: docs/learn/go.md and docs/learn/http.md
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	// Shared code, from packages/go. Learn more: docs/learn/go.md#sharing-code-between-modules
	"github.com/saeventec/reach/packages/go/config"
	"github.com/saeventec/reach/packages/go/database"
	"github.com/saeventec/reach/packages/go/health"
	"github.com/saeventec/reach/packages/go/httpserver"

	"github.com/saeventec/reach/services/auth/gen/proto/auth/v1/authv1connect"
	"github.com/saeventec/reach/services/auth/gen/queries"
)

// maximumRequestBytes is the largest API message the service reads. The
// biggest real one, a sign-up, is under 1 KiB. Without a limit, a client
// could send huge messages that use up memory and time. Bigger messages
// get a RESOURCE_EXHAUSTED error.
const maximumRequestBytes = 16 * 1024

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

	secureCookies, err := readCookieSecure()
	if err != nil {
		return err
	}

	pool, err := database.Connect(stopSignal, databaseURL)
	if err != nil {
		return err
	}
	// Close the database connections last, after the server has stopped.
	defer pool.Close()

	// queries.New wraps the pool with the functions sqlc generated.
	databaseQueries := queries.New(pool)
	users := &postgresUserStore{queries: databaseQueries}
	sessions := &postgresSessionStore{queries: databaseQueries}

	// The rate limits keep their counts in memory. A goroutine forgets old
	// counts every minute, until the stop signal arrives.
	limiters := newLoginLimiters()
	go limiters.removeFullBucketsRegularly(stopSignal)

	port := config.Port()
	address := ":" + port

	server := &http.Server{
		Addr:    address,
		Handler: newRouter(users, sessions, limiters, secureCookies),
		// Close connections that are too slow to send their request headers.
		// Without a limit, a slow client could keep connections open forever.
		ReadHeaderTimeout: 5 * time.Second,
	}

	return httpserver.Run(stopSignal, "auth", server)
}

// newRouter connects each URL path to the code that handles it.
func newRouter(users userStore, sessions sessionStore, limiters *loginLimiters, secureCookies bool) *http.ServeMux {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", health.Handle)

	// The Connect server answers the API calls defined in proto/. Mount adds
	// one path per call to our router, like /auth.v1.AccountService/SignUp.
	accountService := &accountServer{users: users, limiters: limiters}
	connectServer := connect.NewServer()
	authv1connect.RegisterAccountServiceHandler(connectServer, accountService)

	sessionService := &sessionServer{users: users, sessions: sessions, limiters: limiters, secureCookies: secureCookies}
	authv1connect.RegisterSessionServiceHandler(connectServer, sessionService)
	connecthttp.Mount(router, connectServer, connecthttp.WithReadMaxBytes(maximumRequestBytes))

	return router
}
