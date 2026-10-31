package main

// Talking to Postgres with pgx. Learn more: docs/learn/pgx.md

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// connectTimeout is how long the service waits for the database at startup.
const connectTimeout = 5 * time.Second

// connectToDatabase creates a connection pool and checks that the
// database answers.
func connectToDatabase(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create the database pool: %w", err)
	}

	// The pool only opens a connection when it is first needed. Ping now,
	// so a wrong address or password is reported at startup, not on the
	// first request.
	pingContext, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	err = pool.Ping(pingContext)
	if err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to the database: %w", err)
	}
	return pool, nil
}
