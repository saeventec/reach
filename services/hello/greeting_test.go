package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"

	hellov1 "github.com/saeventec/reach/services/hello/gen/hello/v1"
)

// fakeGreetingStore stands in for the database in tests. It returns
// whatever message and error the test gives it.
type fakeGreetingStore struct {
	message string
	err     error
}

func (store *fakeGreetingStore) NewestGreeting(ctx context.Context) (string, error) {
	return store.message, store.err
}

func TestGetGreetingReturnsTheNewestGreeting(t *testing.T) {
	store := &fakeGreetingStore{message: "Hello, test!"}
	server := &greetingServer{greetings: store}

	response, err := server.GetGreeting(t.Context(), &hellov1.GetGreetingRequest{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetMessage() != "Hello, test!" {
		t.Errorf("message: got %q, want %q", response.GetMessage(), "Hello, test!")
	}
}

func TestGetGreetingReportsAMissingGreeting(t *testing.T) {
	store := &fakeGreetingStore{err: errNoGreeting}
	server := &greetingServer{greetings: store}

	_, err := server.GetGreeting(t.Context(), &hellov1.GetGreetingRequest{})

	code := connect.CodeOf(err)
	if code != connect.CodeNotFound {
		t.Errorf("code: got %v, want %v", code, connect.CodeNotFound)
	}
}

func TestGetGreetingHidesDatabaseErrors(t *testing.T) {
	store := &fakeGreetingStore{err: errors.New("connection refused")}
	server := &greetingServer{greetings: store}

	_, err := server.GetGreeting(t.Context(), &hellov1.GetGreetingRequest{})

	code := connect.CodeOf(err)
	if code != connect.CodeInternal {
		t.Errorf("code: got %v, want %v", code, connect.CodeInternal)
	}
	if strings.Contains(err.Error(), "connection refused") {
		t.Errorf("the error sent to the client shows internal details: %q", err.Error())
	}
}

// This test sends a real HTTP request through the router, the way a
// browser would: a POST with a JSON body to the call's path.
func TestGetGreetingOverHTTP(t *testing.T) {
	store := &fakeGreetingStore{message: "Hello, test!"}
	router := newRouter(store)

	requestBody := strings.NewReader("{}")
	request := httptest.NewRequest(http.MethodPost, "/hello.v1.GreetingService/GetGreeting", requestBody)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}

	var responseBody struct {
		Message string `json:"message"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("the response is not valid JSON: %v", err)
	}
	if responseBody.Message != "Hello, test!" {
		t.Errorf("message: got %q, want %q", responseBody.Message, "Hello, test!")
	}
}
