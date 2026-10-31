package main

// Talking to Postgres with pgx. Learn more: docs/learn/pgx.md

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// connectTimeout is how long the service waits for the database at startup.
const connectTimeout = 5 * time.Second

// newestGreetingQuery reads the row with the highest id: the greeting that
// was added last.
const newestGreetingQuery = `SELECT id, message FROM greetings ORDER BY id DESC LIMIT 1`

// errNoGreeting means the greetings table is empty.
var errNoGreeting = errors.New("no greeting found")

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

// postgresGreetingStore reads greetings from the Postgres database.
type postgresGreetingStore struct {
	pool *pgxpool.Pool
}

// NewestGreeting returns the greeting that was added last.
func (store *postgresGreetingStore) NewestGreeting(ctx context.Context) (greeting, error) {
	row := store.pool.QueryRow(ctx, newestGreetingQuery)

	// Scan copies the columns, in the order of the SELECT, into the fields.
	var newest greeting
	err := row.Scan(&newest.id, &newest.message)
	if errors.Is(err, pgx.ErrNoRows) {
		return greeting{}, errNoGreeting
	}
	if err != nil {
		return greeting{}, fmt.Errorf("read the newest greeting: %w", err)
	}
	return newest, nil
}
