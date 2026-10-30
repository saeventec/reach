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

## Alternatives

- **testify:** a popular library with shorter checks (`assert.Equal`). The standard `testing` package is enough for now.
- **Ginkgo:** a test framework in a different style ("describe ... it ..."), more to learn.

## Learn more

- https://go.dev/doc/tutorial/add-a-test
- https://pkg.go.dev/net/http/httptest
