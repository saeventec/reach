package main

// Saving and finding sessions in Postgres. Learn more: docs/learn/sessions.md

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/saeventec/reach/services/auth/gen/queries"
)

// errNoSession means there is no valid session for a token: it never
// existed, or it has expired.
var errNoSession = errors.New("no valid session")

// currentSession is a valid session and the user it belongs to.
type currentSession struct {
	id       string
	userID   string
	username string
}

// sessionStore is what the session server needs from the database.
type sessionStore interface {
	CreateSession(ctx context.Context, userID string, tokenHash []byte, userAgent string, expiresAt time.Time) error
	RenewSession(ctx context.Context, tokenHash []byte, newExpiresAt time.Time) (currentSession, error)
}

// postgresSessionStore keeps the sessions in the auth database.
type postgresSessionStore struct {
	queries *queries.Queries
}

// CreateSession saves a new session for a user, made in the browser that
// sent this User-Agent.
func (store *postgresSessionStore) CreateSession(ctx context.Context, userID string, tokenHash []byte, userAgent string, expiresAt time.Time) error {
	databaseUserID, err := parseUUID(userID)
	if err != nil {
		return err
	}

	parameters := queries.CreateSessionParams{
		UserID:    databaseUserID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
	}
	err = store.queries.CreateSession(ctx, parameters)
	if err != nil {
		return fmt.Errorf("create the session: %w", err)
	}
	return nil
}

// RenewSession finds the valid session with this token hash, moves its
// expiry to newExpiresAt, and returns it with its user. It returns
// errNoSession if there is no valid session.
func (store *postgresSessionStore) RenewSession(ctx context.Context, tokenHash []byte, newExpiresAt time.Time) (currentSession, error) {
	parameters := queries.RenewSessionParams{
		TokenHash:    tokenHash,
		NewExpiresAt: newExpiresAt,
	}
	row, err := store.queries.RenewSession(ctx, parameters)
	if errors.Is(err, pgx.ErrNoRows) {
		return currentSession{}, errNoSession
	}
	if err != nil {
		return currentSession{}, fmt.Errorf("renew the session: %w", err)
	}

	session := currentSession{
		id:       row.SessionID.String(),
		userID:   row.UserID.String(),
		username: row.Username,
	}
	return session, nil
}

// parseUUID turns a UUID in text form into the type the queries use.
func parseUUID(text string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(text)
	if err != nil {
		return pgtype.UUID{}, fmt.Errorf("read the uuid %q: %w", text, err)
	}
	return id, nil
}
