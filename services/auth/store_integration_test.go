package main

// Integration tests: the postgres stores against a real Postgres, started
// in Docker by Testcontainers (ADR 0023). They test what the fakes can't:
// the SQL itself, the migrations, and the database's own rules.
// They need Docker running. "go test -short" skips them.
// Learn more: docs/learn/testing.md#integration-tests

import (
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	// The "_" import only runs the package's setup code: it registers pgx
	// as a driver for Go's database/sql, which goose uses.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/saeventec/reach/services/auth/gen/queries"
)

// testPostgresImage is the same Postgres version as in compose.yaml.
const testPostgresImage = "postgres:18.6"

func TestPostgresStores(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	pool := startTestDatabase(t)
	databaseQueries := queries.New(pool)
	users := &postgresUserStore{queries: databaseQueries}
	sessions := &postgresSessionStore{queries: databaseQueries}

	// Each subtest uses its own usernames, so they can share one database.

	t.Run("creates and finds a user", func(t *testing.T) {
		created, err := users.CreateUser(t.Context(), "alice", "the-hash")
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		found, err := users.FindUserByUsername(t.Context(), "alice")

		if err != nil {
			t.Fatalf("find: %v", err)
		}
		if found.id != created.id || found.passwordHash != "the-hash" {
			t.Errorf("got %+v, want the created user with its hash", found)
		}
	})

	t.Run("refuses a taken username", func(t *testing.T) {
		createTestUser(t, users, "bob")

		_, err := users.CreateUser(t.Context(), "bob", "another-hash")

		if !errors.Is(err, errUsernameTaken) {
			t.Errorf("got %v, want %v", err, errUsernameTaken)
		}
	})

	t.Run("the database refuses a username that breaks the rules", func(t *testing.T) {
		// The Go code checks the rules first; the CHECK constraint is the
		// safety net, in case a bug lets a bad name through.
		_, err := users.CreateUser(t.Context(), "Not Valid", "the-hash")

		if err == nil {
			t.Fatalf("a bad username was saved")
		}
		if errors.Is(err, errUsernameTaken) {
			t.Errorf("a broken rule was reported as a taken username")
		}
	})

	t.Run("reports an unknown username", func(t *testing.T) {
		_, err := users.FindUserByUsername(t.Context(), "nobody")

		if !errors.Is(err, errUserNotFound) {
			t.Errorf("got %v, want %v", err, errUserNotFound)
		}
	})

	t.Run("renews a valid session", func(t *testing.T) {
		user := createTestUser(t, users, "carol")
		tokenHash := hashSessionToken("carol-token")
		createTestSession(t, sessions, user.id, tokenHash, time.Now().Add(time.Hour))

		session, err := sessions.RenewSession(t.Context(), tokenHash, time.Now().Add(sessionLifetime))

		if err != nil {
			t.Fatalf("renew: %v", err)
		}
		if session.userID != user.id || session.username != "carol" {
			t.Errorf("got %+v, want carol's session", session)
		}
	})

	t.Run("refuses an expired session", func(t *testing.T) {
		user := createTestUser(t, users, "dave")
		tokenHash := hashSessionToken("dave-token")
		createTestSession(t, sessions, user.id, tokenHash, time.Now().Add(-time.Minute))

		_, err := sessions.RenewSession(t.Context(), tokenHash, time.Now().Add(sessionLifetime))

		if !errors.Is(err, errNoSession) {
			t.Errorf("got %v, want %v", err, errNoSession)
		}
	})

	t.Run("lists only valid sessions, the most recently used first", func(t *testing.T) {
		user := createTestUser(t, users, "erin")
		createTestSession(t, sessions, user.id, hashSessionToken("erin-old"), time.Now().Add(time.Hour))
		createTestSession(t, sessions, user.id, hashSessionToken("erin-new"), time.Now().Add(time.Hour))
		createTestSession(t, sessions, user.id, hashSessionToken("erin-expired"), time.Now().Add(-time.Minute))
		// Use the second session, so it is the most recently used.
		_, err := sessions.RenewSession(t.Context(), hashSessionToken("erin-new"), time.Now().Add(sessionLifetime))
		if err != nil {
			t.Fatalf("renew: %v", err)
		}

		listed, err := sessions.ListSessions(t.Context(), user.id)

		if err != nil {
			t.Fatalf("list: %v", err)
		}
		if len(listed) != 2 {
			t.Fatalf("got %d sessions, want 2 (the expired one left out)", len(listed))
		}
		if !listed[0].lastUsedAt.After(listed[1].lastUsedAt) {
			t.Errorf("the most recently used session is not first")
		}
	})

	t.Run("a user can't delete another user's session", func(t *testing.T) {
		owner := createTestUser(t, users, "frank")
		stranger := createTestUser(t, users, "grace")
		createTestSession(t, sessions, owner.id, hashSessionToken("frank-token"), time.Now().Add(time.Hour))
		frankSessions, err := sessions.ListSessions(t.Context(), owner.id)
		if err != nil || len(frankSessions) != 1 {
			t.Fatalf("list: got %d sessions, error %v", len(frankSessions), err)
		}
		sessionID := frankSessions[0].id

		err = sessions.DeleteSession(t.Context(), sessionID, stranger.id)

		if !errors.Is(err, errNoSession) {
			t.Errorf("got %v, want %v", err, errNoSession)
		}
		// The session must still be there and work.
		_, err = sessions.RenewSession(t.Context(), hashSessionToken("frank-token"), time.Now().Add(sessionLifetime))
		if err != nil {
			t.Errorf("frank's session was deleted by another user: %v", err)
		}
	})

	t.Run("a user can delete their own session", func(t *testing.T) {
		user := createTestUser(t, users, "heidi")
		createTestSession(t, sessions, user.id, hashSessionToken("heidi-token"), time.Now().Add(time.Hour))
		listed, err := sessions.ListSessions(t.Context(), user.id)
		if err != nil || len(listed) != 1 {
			t.Fatalf("list: got %d sessions, error %v", len(listed), err)
		}

		err = sessions.DeleteSession(t.Context(), listed[0].id, user.id)

		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		_, err = sessions.RenewSession(t.Context(), hashSessionToken("heidi-token"), time.Now().Add(sessionLifetime))
		if !errors.Is(err, errNoSession) {
			t.Errorf("the deleted session still works: %v", err)
		}
	})

	t.Run("an id that is not a UUID deletes nothing", func(t *testing.T) {
		user := createTestUser(t, users, "ivan")

		err := sessions.DeleteSession(t.Context(), "not-a-uuid", user.id)

		if !errors.Is(err, errNoSession) {
			t.Errorf("got %v, want %v", err, errNoSession)
		}
	})
}

// startTestDatabase starts a new, empty Postgres in Docker, applies the
// migrations, and returns a pool connected to it. The container is removed
// when the test ends.
func startTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	container, err := postgres.Run(t.Context(), testPostgresImage,
		postgres.WithDatabase("auth"),
		postgres.WithUsername("auth"),
		postgres.WithPassword("test-password"),
		// Wait until Postgres accepts connections.
		postgres.BasicWaitStrategies(),
	)
	// Schedule the removal first, so the container is removed even if a
	// later step fails.
	testcontainers.CleanupContainer(t, container)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}

	databaseURL, err := container.ConnectionString(t.Context(), "sslmode=disable")
	if err != nil {
		t.Fatalf("read the database address: %v", err)
	}

	applyMigrations(t, databaseURL)

	pool, err := pgxpool.New(t.Context(), databaseURL)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// applyMigrations runs the files in migrations/ with goose, the same tool
// (and version) that the auth-migrate container uses.
func applyMigrations(t *testing.T, databaseURL string) {
	t.Helper()

	// goose works with Go's general database/sql package, so it gets its
	// own connection, closed when the test ends.
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open the database for goose: %v", err)
	}
	t.Cleanup(func() {
		closeErr := database.Close()
		if closeErr != nil {
			t.Errorf("close the goose connection: %v", closeErr)
		}
	})

	migrationFiles := os.DirFS("migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, migrationFiles)
	if err != nil {
		t.Fatalf("prepare the migrations: %v", err)
	}

	_, err = provider.Up(t.Context())
	if err != nil {
		t.Fatalf("apply the migrations: %v", err)
	}
}

// createTestUser saves a user, and fails the test if that doesn't work.
func createTestUser(t *testing.T, users *postgresUserStore, username string) createdUser {
	t.Helper()
	user, err := users.CreateUser(t.Context(), username, "the-hash")
	if err != nil {
		t.Fatalf("create the user %q: %v", username, err)
	}
	return user
}

// createTestSession saves a session, and fails the test if that doesn't work.
func createTestSession(t *testing.T, sessions *postgresSessionStore, userID string, tokenHash []byte, expiresAt time.Time) {
	t.Helper()
	err := sessions.CreateSession(t.Context(), userID, tokenHash, "test agent", expiresAt)
	if err != nil {
		t.Fatalf("create the session: %v", err)
	}
}
