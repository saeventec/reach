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

	"github.com/jackc/pgx/v5/pgconn"
)

// fakeUserStore stands in for the database in tests. It remembers what it
// was asked to save, and returns the user and the error the test gives it.
type fakeUserStore struct {
	err error

	savedUsername     string
	savedPasswordHash string

	// foundUser is what FindUserByUsername returns when err is nil.
	foundUser loginUser

	// What the directory's lookups were asked for, and what they return.
	lookedUpUsernames []string
	lookedUpUserIDs   []string
	foundUsers        []publicUser
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

func (store *fakeUserStore) FindUsersByUsernames(ctx context.Context, usernames []string) ([]publicUser, error) {
	if store.err != nil {
		return nil, store.err
	}
	store.lookedUpUsernames = usernames
	return store.foundUsers, nil
}

func (store *fakeUserStore) FindUsersByIDs(ctx context.Context, userIDs []string) ([]publicUser, error) {
	if store.err != nil {
		return nil, store.err
	}
	store.lookedUpUserIDs = userIDs
	return store.foundUsers, nil
}

// testUserID is the id the fake store gives every new account.
const testUserID = "019a3f2c-7b1e-7c4d-9a2b-3c4d5e6f7a8b"

// signUpPath is the address of the SignUp call.
const signUpPath = "/auth.v1.AccountService/SignUp"

// credentialsBody is the JSON body of a sign-up or a login with this
// username and password.
func credentialsBody(username string, password string) string {
	body, err := json.Marshal(map[string]string{"username": username, "password": password})
	if err != nil {
		panic(err) // a map of strings can always be turned into JSON
	}
	return string(body)
}

// errorCode reads the Connect error code from a JSON error response, like
// "invalid_argument".
func errorCode(t *testing.T, recorder *httptest.ResponseRecorder) string {
	t.Helper()
	var responseBody struct {
		Code string `json:"code"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("the response is not valid JSON: %v (body: %s)", err, recorder.Body.String())
	}
	return responseBody.Code
}

func TestSignUpCreatesTheAccount(t *testing.T) {
	store := &fakeUserStore{}
	router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	recorder := callAPI(router, signUpPath, credentialsBody(" Sadra ", "purple tram winter"), "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	// The username is stored in its normalized form.
	if store.savedUsername != "sadra" {
		t.Errorf("saved username: got %q, want %q", store.savedUsername, "sadra")
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
		router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

		recorder := callAPI(router, signUpPath, credentialsBody(testCase.username, testCase.password), "")

		code := errorCode(t, recorder)
		if code != "invalid_argument" {
			t.Errorf("SignUp(%q, %q): code: got %q, want %q", testCase.username, testCase.password, code, "invalid_argument")
		}
		if store.savedUsername != "" {
			t.Errorf("SignUp(%q, %q): the account was saved anyway", testCase.username, testCase.password)
		}
	}
}

func TestSignUpReportsATakenUsername(t *testing.T) {
	store := &fakeUserStore{err: errUsernameTaken}
	router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	recorder := callAPI(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "")

	code := errorCode(t, recorder)
	if code != "already_exists" {
		t.Errorf("code: got %q, want %q", code, "already_exists")
	}
}

func TestSignUpHidesDatabaseErrors(t *testing.T) {
	store := &fakeUserStore{err: errors.New("connection refused")}
	router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	recorder := callAPI(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "")

	code := errorCode(t, recorder)
	if code != "internal" {
		t.Errorf("code: got %q, want %q", code, "internal")
	}
	if strings.Contains(recorder.Body.String(), "connection refused") {
		t.Errorf("the error sent to the client shows internal details: %s", recorder.Body.String())
	}
}

// This test sends a real HTTP request through the router, the way the
// browser will: a POST with a JSON body to the call's path.
func TestSignUpOverHTTP(t *testing.T) {
	store := &fakeUserStore{}
	router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

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
	router := newRouter(store, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)
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
