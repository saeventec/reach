package main

// Reading greetings from Postgres with pgx. Connecting is shared by all
// services: see packages/go/database. Learn more: docs/learn/pgx.md

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// newestGreetingQuery reads the row with the highest id: the greeting that
// was added last.
const newestGreetingQuery = `SELECT id, message FROM greetings ORDER BY id DESC LIMIT 1`

// errNoGreeting means the greetings table is empty.
var errNoGreeting = errors.New("no greeting found")

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
