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

## Alternatives

- **testify:** a popular library with shorter checks (`assert.Equal`). The standard `testing` package is enough for now.
- **Ginkgo:** a test framework in a different style ("describe ... it ..."), more to learn.

## Learn more

- https://go.dev/doc/tutorial/add-a-test
- https://pkg.go.dev/net/http/httptest
