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
// whatever greeting and error the test gives it.
type fakeGreetingStore struct {
	greeting greeting
	err      error
}

func (store *fakeGreetingStore) NewestGreeting(ctx context.Context) (greeting, error) {
	return store.greeting, store.err
}

// testGreeting is the greeting the fake store returns in these tests.
var testGreeting = greeting{id: 1, message: "Hello, test!"}

func TestGetGreetingReturnsTheNewestGreeting(t *testing.T) {
	store := &fakeGreetingStore{greeting: testGreeting}
	server := &greetingServer{greetings: store}

	response, err := server.GetGreeting(t.Context(), &hellov1.GetGreetingRequest{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetMessage() != testGreeting.message {
		t.Errorf("message: got %q, want %q", response.GetMessage(), testGreeting.message)
	}
	if response.GetId() != testGreeting.id {
		t.Errorf("id: got %d, want %d", response.GetId(), testGreeting.id)
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
	store := &fakeGreetingStore{greeting: testGreeting}
	router := newRouter(store)

	requestBody := strings.NewReader("{}")
	request := httptest.NewRequest(http.MethodPost, "/hello.v1.GreetingService/GetGreeting", requestBody)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}

	// In JSON, Protobuf sends int64 numbers as strings ("id":"1"), because
	// JavaScript can't hold such big numbers exactly. ",string" reads them.
	var responseBody struct {
		Message string `json:"message"`
		ID      int64  `json:"id,string"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("the response is not valid JSON: %v", err)
	}
	if responseBody.Message != testGreeting.message {
		t.Errorf("message: got %q, want %q", responseBody.Message, testGreeting.message)
	}
	if responseBody.ID != testGreeting.id {
		t.Errorf("id: got %d, want %d", responseBody.ID, testGreeting.id)
	}
}
