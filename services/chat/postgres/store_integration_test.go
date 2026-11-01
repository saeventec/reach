package postgres

// Integration tests: the store against a real Postgres, started in Docker
// by Testcontainers (ADR 0023). They test what the fakes can't: the SQL,
// the migrations, the transactions and the database's own rules.
// They need Docker running. "go test -short" skips them.
// Learn more: docs/learn/testing.md#integration-tests

import (
	"database/sql"
	"os"
	"slices"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	// The "_" import only runs the package's setup code: it registers pgx
	// as a driver for Go's database/sql, which goose uses.
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	testpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/saeventec/reach/services/chat/conversation"
)

// testPostgresImage is the same Postgres version as in compose.yaml.
const testPostgresImage = "postgres:18.6"

// Made-up user ids. The chat database doesn't know users, only their ids.
const (
	aliceID = "01a00000-0000-7000-8000-00000000000a"
	bobID   = "01a00000-0000-7000-8000-00000000000b"
	carolID = "01a00000-0000-7000-8000-00000000000c"
	daveID  = "01a00000-0000-7000-8000-00000000000d"
)

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test: needs Docker")
	}

	pool := startTestDatabase(t)
	store := NewStore(pool)

	// Each subtest uses its own people, so they can share one database.

	t.Run("finds the same 1:1 conversation from both sides", func(t *testing.T) {
		first, err := store.FindOrCreateDirectConversation(t.Context(), "alice:bob", aliceID, bobID)
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		second, err := store.FindOrCreateDirectConversation(t.Context(), "alice:bob", bobID, aliceID)
		if err != nil {
			t.Fatalf("find: %v", err)
		}

		if first.ID != second.ID {
			t.Errorf("got two conversations (%s, %s), want one", first.ID, second.ID)
		}
		if first.Kind != conversation.KindDirect || len(first.Members) != 2 {
			t.Errorf("got %+v, want a direct conversation with 2 members", first)
		}
	})

	t.Run("two requests at the same moment still make one 1:1 conversation", func(t *testing.T) {
		// Ten goroutines (functions running at the same time) all try to
		// create the same pair's conversation. The UNIQUE pair key and
		// ON CONFLICT make sure only one is created.
		var waitGroup sync.WaitGroup
		ids := make([]string, 10)
		errs := make([]error, 10)
		for number := range 10 {
			waitGroup.Go(func() {
				found, err := store.FindOrCreateDirectConversation(t.Context(), "carol:dave", carolID, daveID)
				ids[number] = found.ID
				errs[number] = err
			})
		}
		waitGroup.Wait()

		for number := range 10 {
			if errs[number] != nil {
				t.Fatalf("request %d: %v", number, errs[number])
			}
			if ids[number] != ids[0] {
				t.Errorf("request %d got %s, request 0 got %s: want one conversation", number, ids[number], ids[0])
			}
		}
	})

	t.Run("creates a group with its members", func(t *testing.T) {
		group, err := store.CreateGroupConversation(t.Context(), "Friends", []string{aliceID, bobID, carolID})

		if err != nil {
			t.Fatalf("create: %v", err)
		}
		if group.Kind != conversation.KindGroup || group.GroupName != "Friends" {
			t.Errorf("got %+v, want a group named Friends", group)
		}
		memberIDs := memberIDsOf(group)
		if !slices.Equal(memberIDs, []string{aliceID, bobID, carolID}) {
			t.Errorf("members: got %v, want alice, bob and carol in that order", memberIDs)
		}
	})

	t.Run("lists only a user's own conversations, the newest first", func(t *testing.T) {
		erin := "01a00000-0000-7000-8000-00000000000e"
		frank := "01a00000-0000-7000-8000-00000000000f"
		older, err := store.CreateGroupConversation(t.Context(), "Older", []string{erin, frank})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		newer, err := store.CreateGroupConversation(t.Context(), "Newer", []string{erin, frank})
		if err != nil {
			t.Fatalf("create: %v", err)
		}
		// A conversation erin is not in.
		_, err = store.CreateGroupConversation(t.Context(), "Not erin's", []string{frank, aliceID})
		if err != nil {
			t.Fatalf("create: %v", err)
		}

		conversations, err := store.ListConversationsOfUser(t.Context(), erin)

		if err != nil {
			t.Fatalf("list: %v", err)
		}
		ids := []string{}
		for _, found := range conversations {
			ids = append(ids, found.ID)
		}
		if !slices.Equal(ids, []string{newer.ID, older.ID}) {
			t.Errorf("got %v, want [newer older]", ids)
		}
		if len(conversations[0].Members) != 2 {
			t.Errorf("got %d members, want 2", len(conversations[0].Members))
		}
	})

	t.Run("the database refuses a group without a name", func(t *testing.T) {
		// The Go code checks the rules first; the CHECK constraint is the
		// safety net, in case a bug lets a bad group through.
		_, err := pool.Exec(t.Context(), "INSERT INTO conversations (kind) VALUES ('group')")

		if err == nil {
			t.Error("a group without a name was saved")
		}
	})
}

// memberIDsOf lists a conversation's member ids, in order.
func memberIDsOf(found conversation.Conversation) []string {
	ids := []string{}
	for _, member := range found.Members {
		ids = append(ids, member.UserID)
	}
	return ids
}

// startTestDatabase starts a new, empty Postgres in Docker, applies the
// migrations, and returns a pool connected to it. The container is removed
// when the test ends. (The same as in the auth service's tests.)
func startTestDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	container, err := testpostgres.Run(t.Context(), testPostgresImage,
		testpostgres.WithDatabase("chat"),
		testpostgres.WithUsername("chat"),
		testpostgres.WithPassword("test-password"),
		// Wait until Postgres accepts connections.
		testpostgres.BasicWaitStrategies(),
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
// the chat-migrate container uses.
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

	// Tests run inside their package's folder (postgres/), so the
	// migrations are one folder up.
	migrationFiles := os.DirFS("../migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, migrationFiles)
	if err != nil {
		t.Fatalf("prepare the migrations: %v", err)
	}

	_, err = provider.Up(t.Context())
	if err != nil {
		t.Fatalf("apply the migrations: %v", err)
	}
}
