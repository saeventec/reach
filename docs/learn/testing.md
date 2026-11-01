# Testing

## What it is

A test is a small program that runs your code and checks the result. If someone breaks the code later, the test fails and tells them right away.

## Why we use it

Tests from day one: each change can be checked in seconds, on your computer and later in CI.

## How Go tests work

- Tests live next to the code, in files ending with `_test.go`.
- Each test is a function named `TestSomething` that takes `t *testing.T`.
- `t.Errorf(...)` marks the test as failed, with a message, and keeps going.
- `httptest` (part of Go) creates fake requests and records the response, so we can test a handler without starting a real server or using the network.

## Tiny example

From `services/hello/health_test.go`:

```go
func TestHealthAnswersOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil) // a fake request
	recorder := httptest.NewRecorder()                               // catches the response

	router := newRouter()
	router.ServeHTTP(recorder, request) // send the request through our router

	if recorder.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
}
```

Run all tests:

```sh
task test
```

## Fakes

A unit test should not need a real database: that would make it slow and fragile. So the code asks for an **interface** (a list of methods) instead of the real database type, and the test passes in a **fake**: a tiny stand-in that returns whatever the test wants.

From `services/hello`:

```go
// greeting.go: the server only needs "something with a NewestGreeting method".
type greetingStore interface {
	NewestGreeting(ctx context.Context) (string, error)
}

// greeting_test.go: a fake that returns what the test tells it to.
type fakeGreetingStore struct {
	message string
	err     error
}

func (store *fakeGreetingStore) NewestGreeting(ctx context.Context) (string, error) {
	return store.message, store.err
}

// In a test: pretend the database is down.
store := &fakeGreetingStore{err: errors.New("connection refused")}
```

With fakes, we can test cases that are hard to create for real, like a database outage. The real database code is tested with **integration tests** against a real Postgres (see [Integration tests](#integration-tests) below).

`t.Context()` gives each test a context that is cancelled when the test ends. `t.Fatalf` stops the test at once, which is useful when the next checks make no sense after a failure.

## Table-driven tests

When the same check runs on many inputs, write the inputs as a list (a "table") and loop over it. Adding a case is one line. From [services/auth/username_test.go](../../services/auth/username_test.go):

```go
testCases := []struct {
	username string
	want     error
}{
	{"sadra", nil},                         // allowed
	{"ab", errUsernameLength},              // too short
	{"Sadra", errUsernameCharacters},       // uppercase
}

for _, testCase := range testCases {
	got := validateUsername(testCase.username)
	if !errors.Is(got, testCase.want) {
		t.Errorf("validateUsername(%q): got %v, want %v", testCase.username, got, testCase.want)
	}
}
```

- `[]struct{...}{...}` is a list of small, unnamed structs: one per case.
- `errors.Is(got, want)` checks that the error is the expected one. `errors.Is(nil, nil)` is true, so "no error" cases work too.
- The message names the input, so a failure says exactly which case broke.

## Integration tests

A **unit test** checks one piece on its own, with fakes for everything around it. It is fast, but a fake store never runs our SQL. An **integration test** checks pieces working together for real: here, the Postgres stores against a real Postgres.

[services/auth/store_integration_test.go](../../services/auth/store_integration_test.go) uses **Testcontainers** ([ADR 0023](../decisions/0023-use-testcontainers-for-integration-tests.md)): the test itself starts a new, empty Postgres in Docker, applies our migrations with goose (the same tool and version as `auth-migrate`), runs the cases, and removes the container at the end.

```go
container, err := postgres.Run(t.Context(), "postgres:18.6",
	postgres.WithDatabase("auth"),
	postgres.BasicWaitStrategies(),          // wait until Postgres accepts connections
)
testcontainers.CleanupContainer(t, container) // remove it when the test ends
```

- **One container, many cases:** starting Postgres takes a few seconds, so `TestPostgresStores` starts it once, and each case is a **subtest** (`t.Run("refuses an expired session", ...)`). Each case uses its own usernames, so they don't get in each other's way.
- **They need Docker running.** `go test -short ./...` skips them (`testing.Short()`), for a quick run without Docker. `task test` and CI run them.
- **NATS too.** [services/chat/natspublisher/publisher_integration_test.go](../../services/chat/natspublisher/publisher_integration_test.go) starts a NATS server the same way, with Testcontainers' NATS module (`testnats.Run`).
- **They test what fakes can't:** the SQL (an expired session is refused, a user can't delete another user's session), the migrations, and the database's own rules (`UNIQUE`, `CHECK`).
- **Break it on purpose:** a test is only useful if it can fail. Removing `AND user_id = ...` from `DeleteSession` makes "a user can't delete another user's session" fail with "frank's session was deleted by another user". Try it yourself, then put the line back.

**The cost:** Testcontainers brings Docker's Go client and many other modules into `go.mod` (about 130 more). They are only used by the tests, never compiled into the service.

## Web app tests (Vitest)

The web app uses:

- **Vitest:** the test runner made for Vite. It runs `*.test.tsx` files.
- **jsdom:** a fake browser inside Node.js, so components can be drawn without a real browser. A comment at the top of the test file turns it on: `// @vitest-environment jsdom`.
- **Testing Library:** draws a component and finds things on the page **the way a user does**: by the text they see, not by internal details.

From `apps/web/src/greeting/NewestGreeting.test.tsx`:

```tsx
test("shows the greeting once it has loaded", async () => {
  const loadGreeting = async () => ({ id: 2n, message: "Hello, test!" });   // a fake

  render(<NewestGreeting loadGreeting={loadGreeting} />);

  screen.getByText("Loading the greeting…");    // must be on the page now
  await screen.findByText("Hello, test!");      // waits for it to appear
});
```

- `2n` is a `bigint` literal (the `n` at the end).
- `vi.spyOn(console, "error")` replaces `console.error` with a stand-in, so a test can check that an error was logged.
- `cleanup()` after each test removes what it drew.

`task test` runs the Go tests and the web tests. `task test:e2e` runs the end-to-end tests in a real browser, against a fresh copy of the whole system (see [playwright.md](playwright.md)). `task test:race` runs the tests of `realtime` and `chat` with Go's race detector (it needs a C compiler; see [websocket.md](websocket.md#the-race-detector)).

## Alternatives

- **testify:** a popular library with shorter checks (`assert.Equal`). The standard `testing` package is enough for now.
- **Ginkgo:** a test framework in a different style ("describe ... it ..."), more to learn.
- **Jest (instead of Vitest):** the older standard for JavaScript tests. Vitest works the same way, but uses Vite's own setup, so it needs no extra config.
- **Playwright:** tests in a real browser. We add it in milestone 3, for whole user flows.

## Learn more

- https://go.dev/doc/tutorial/add-a-test
- https://pkg.go.dev/net/http/httptest
