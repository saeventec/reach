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

With fakes, we can test cases that are hard to create for real, like a database outage. The real database code is tested later with **integration tests** against a real Postgres (Testcontainers, milestone 2).

`t.Context()` gives each test a context that is cancelled when the test ends. `t.Fatalf` stops the test at once, which is useful when the next checks make no sense after a failure.

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

`task test` runs the Go tests and the web tests.

## Alternatives

- **testify:** a popular library with shorter checks (`assert.Equal`). The standard `testing` package is enough for now.
- **Ginkgo:** a test framework in a different style ("describe ... it ..."), more to learn.
- **Jest (instead of Vitest):** the older standard for JavaScript tests. Vitest works the same way, but uses Vite's own setup, so it needs no extra config.
- **Playwright:** tests in a real browser. We add it in milestone 3, for whole user flows.

## Learn more

- https://go.dev/doc/tutorial/add-a-test
- https://pkg.go.dev/net/http/httptest
