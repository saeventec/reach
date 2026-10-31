package main

// Saving and reading accounts in Postgres, with the queries that sqlc
// generates from queries/users.sql. Learn more: docs/learn/sqlc.md

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/saeventec/reach/services/auth/gen/queries"
)

// uniqueViolationCode is the Postgres error code for "this value already
// exists" (a UNIQUE rule was broken). Every Postgres error has a 5-character
// code: https://www.postgresql.org/docs/current/errcodes-appendix.html
const uniqueViolationCode = "23505"

// errUsernameTaken means another account already has this username.
var errUsernameTaken = errors.New("the username is already taken")

// errUserNotFound means no account has this username.
var errUserNotFound = errors.New("no user with this username")

// createdUser is a new account, as the rest of the service sees it.
type createdUser struct {
	id       string
	username string
}

// loginUser is an account with its password hash, as login needs it.
type loginUser struct {
	id           string
	username     string
	passwordHash string
}

// userStore is what the API needs from the users table. It is an
// interface, so tests can use a fake store instead of a real database.
// Learn more: docs/learn/testing.md#fakes
type userStore interface {
	CreateUser(ctx context.Context, username string, passwordHash string) (createdUser, error)
	FindUserByUsername(ctx context.Context, username string) (loginUser, error)
}

// postgresUserStore keeps the accounts in the auth database.
type postgresUserStore struct {
	queries *queries.Queries
}

// CreateUser saves a new account. It returns errUsernameTaken if the
// username is already used.
func (store *postgresUserStore) CreateUser(ctx context.Context, username string, passwordHash string) (createdUser, error) {
	parameters := queries.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
	}
	row, err := store.queries.CreateUser(ctx, parameters)
	if isUniqueViolation(err) {
		return createdUser{}, errUsernameTaken
	}
	if err != nil {
		return createdUser{}, fmt.Errorf("create the user: %w", err)
	}

	user := createdUser{
		id:       row.ID.String(),
		username: row.Username,
	}
	return user, nil
}

// FindUserByUsername returns the account with this username, or
// errUserNotFound.
func (store *postgresUserStore) FindUserByUsername(ctx context.Context, username string) (loginUser, error) {
	row, err := store.queries.FindUserByUsername(ctx, username)
	if errors.Is(err, pgx.ErrNoRows) {
		return loginUser{}, errUserNotFound
	}
	if err != nil {
		return loginUser{}, fmt.Errorf("find the user: %w", err)
	}

	user := loginUser{
		id:           row.ID.String(),
		username:     row.Username,
		passwordHash: row.PasswordHash,
	}
	return user, nil
}

// isUniqueViolation reports whether err is Postgres saying "this value
// already exists". We let the database decide, instead of first asking "is
// the name free?": between that question and the insert, someone else could
// take the name.
func isUniqueViolation(err error) bool {
	// errors.As looks through the wrapped errors for a *pgconn.PgError, the
	// type pgx uses for errors sent by Postgres.
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return false
	}
	return postgresError.Code == uniqueViolationCode
}
