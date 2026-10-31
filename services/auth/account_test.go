package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"connectrpc.com/connect/v2"
	"github.com/jackc/pgx/v5/pgconn"

	authv1 "github.com/saeventec/reach/services/auth/gen/proto/auth/v1"
)

// fakeUserStore stands in for the database in tests. It remembers what it
// was asked to save, and returns the user and the error the test gives it.
type fakeUserStore struct {
	err error

	savedUsername     string
	savedPasswordHash string

	// foundUser is what FindUserByUsername returns when err is nil.
	foundUser loginUser
}

func (store *fakeUserStore) CreateUser(ctx context.Context, username string, passwordHash string) (createdUser, error) {
	if store.err != nil {
		return createdUser{}, store.err
	}
	store.savedUsername = username
	store.savedPasswordHash = passwordHash

	user := createdUser{id: testUserID, username: username}
	return user, nil
}

func (store *fakeUserStore) FindUserByUsername(ctx context.Context, username string) (loginUser, error) {
	if store.err != nil {
		return loginUser{}, store.err
	}
	return store.foundUser, nil
}

// testUserID is the id the fake store gives every new account.
const testUserID = "019a3f2c-7b1e-7c4d-9a2b-3c4d5e6f7a8b"

func TestSignUpCreatesTheAccount(t *testing.T) {
	store := &fakeUserStore{}
	server := &accountServer{users: store}
	request := &authv1.SignUpRequest{Username: " Sadra ", Password: "purple tram winter"}

	response, err := server.SignUp(t.Context(), request)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if response.GetUserId() != testUserID {
		t.Errorf("user id: got %q, want %q", response.GetUserId(), testUserID)
	}
	// The username is stored and returned in its normalized form.
	if store.savedUsername != "sadra" {
		t.Errorf("saved username: got %q, want %q", store.savedUsername, "sadra")
	}
	if response.GetUsername() != "sadra" {
		t.Errorf("username: got %q, want %q", response.GetUsername(), "sadra")
	}
	// The store gets a hash, never the password.
	if !strings.HasPrefix(store.savedPasswordHash, "$argon2id$") {
		t.Errorf("saved password hash: got %q, want an argon2id hash", store.savedPasswordHash)
	}
}

func TestSignUpRejectsBadInput(t *testing.T) {
	testCases := []struct {
		username string
		password string
	}{
		{"ab", "purple tram winter"},        // username too short
		{"sa..dra", "purple tram winter"},   // two separators in a row
		{"sadra", "short"},                  // password too short
		{"Sadra.Samadi", "SADRA.SAMADI"},    // password is the username
		{"sadra", "my reach password"},      // password contains "reach"
		{"sadra", strings.Repeat("a", 129)}, // password too long
	}

	for _, testCase := range testCases {
		store := &fakeUserStore{}
		server := &accountServer{users: store}
		request := &authv1.SignUpRequest{Username: testCase.username, Password: testCase.password}

		_, err := server.SignUp(t.Context(), request)

		code := connect.CodeOf(err)
		if code != connect.CodeInvalidArgument {
			t.Errorf("SignUp(%q, %q): code: got %v, want %v", testCase.username, testCase.password, code, connect.CodeInvalidArgument)
		}
		if store.savedUsername != "" {
			t.Errorf("SignUp(%q, %q): the account was saved anyway", testCase.username, testCase.password)
		}
	}
}

func TestSignUpReportsATakenUsername(t *testing.T) {
	store := &fakeUserStore{err: errUsernameTaken}
	server := &accountServer{users: store}
	request := &authv1.SignUpRequest{Username: "sadra", Password: "purple tram winter"}

	_, err := server.SignUp(t.Context(), request)

	code := connect.CodeOf(err)
	if code != connect.CodeAlreadyExists {
		t.Errorf("code: got %v, want %v", code, connect.CodeAlreadyExists)
	}
}

func TestSignUpHidesDatabaseErrors(t *testing.T) {
	store := &fakeUserStore{err: errors.New("connection refused")}
	server := &accountServer{users: store}
	request := &authv1.SignUpRequest{Username: "sadra", Password: "purple tram winter"}

	_, err := server.SignUp(t.Context(), request)

	code := connect.CodeOf(err)
	if code != connect.CodeInternal {
		t.Errorf("code: got %v, want %v", code, connect.CodeInternal)
	}
	if strings.Contains(err.Error(), "connection refused") {
		t.Errorf("the error sent to the client shows internal details: %q", err.Error())
	}
}

// This test sends a real HTTP request through the router, the way the
// browser will: a POST with a JSON body to the call's path.
func TestSignUpOverHTTP(t *testing.T) {
	store := &fakeUserStore{}
	router := newRouter(store, &fakeSessionStore{}, true)

	requestBody := strings.NewReader(`{"username": "sadra", "password": "purple tram winter"}`)
	request := httptest.NewRequest(http.MethodPost, "/auth.v1.AccountService/SignUp", requestBody)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	// Protobuf's JSON uses lowerCamelCase names: user_id becomes userId.
	var responseBody struct {
		UserID   string `json:"userId"`
		Username string `json:"username"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("the response is not valid JSON: %v", err)
	}
	if responseBody.UserID != testUserID {
		t.Errorf("user id: got %q, want %q", responseBody.UserID, testUserID)
	}
	if responseBody.Username != "sadra" {
		t.Errorf("username: got %q, want %q", responseBody.Username, "sadra")
	}
}

func TestIsUniqueViolation(t *testing.T) {
	uniqueViolation := &pgconn.PgError{Code: uniqueViolationCode}
	// fmt.Errorf with %w wraps an error, as our code does. errors.As must
	// still find the Postgres error inside.
	wrapped := fmt.Errorf("create the user: %w", uniqueViolation)
	otherPostgresError := &pgconn.PgError{Code: "23514"} // a CHECK rule was broken

	if !isUniqueViolation(wrapped) {
		t.Errorf("a wrapped unique violation was not recognized")
	}
	if isUniqueViolation(otherPostgresError) {
		t.Errorf("another Postgres error was taken for a unique violation")
	}
	if isUniqueViolation(errors.New("connection refused")) {
		t.Errorf("a non-Postgres error was taken for a unique violation")
	}
}

func TestSignUpRefusesHugeRequests(t *testing.T) {
	store := &fakeUserStore{}
	router := newRouter(store, &fakeSessionStore{}, true)
	hugePassword := strings.Repeat("a", 20*1024) // 20 KiB, over the 16 KiB limit
	body := `{"username": "sadra", "password": "` + hugePassword + `"}`

	recorder := callAPI(router, "/auth.v1.AccountService/SignUp", body, "")

	if recorder.Code == http.StatusOK {
		t.Fatalf("a huge request was accepted")
	}
	if !strings.Contains(recorder.Body.String(), "resource_exhausted") {
		t.Errorf("body: got %s, want the code resource_exhausted", recorder.Body.String())
	}
	if store.savedUsername != "" {
		t.Errorf("the account was saved anyway")
	}
}
