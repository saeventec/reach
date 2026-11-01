package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
)

const (
	findUsersByUsernamePath = "/authinternal.v1.UserDirectoryService/FindUsersByUsername"
	findUsersByIDPath       = "/authinternal.v1.UserDirectoryService/FindUsersById"
)

func TestFindUsersByUsernameNormalizesTheNames(t *testing.T) {
	users := &fakeUserStore{foundUsers: []publicUser{{id: testUserID, username: "sadra"}}}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	response := callAPI(router, findUsersByUsernamePath, `{"usernames": [" Sadra "]}`, "")

	if response.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (%s)", response.Code, http.StatusOK, response.Body)
	}
	// The same rule as at login: spaces removed, lowercase.
	if !slices.Equal(users.lookedUpUsernames, []string{"sadra"}) {
		t.Errorf("looked up %q, want [sadra]", users.lookedUpUsernames)
	}
	wantBody := fmt.Sprintf(`{"users":[{"id":"%s","username":"sadra"}]}`, testUserID)
	if compactJSON(t, response.Body.String()) != wantBody {
		t.Errorf("body: got %s, want %s", response.Body, wantBody)
	}
}

func TestFindUsersByUsernameRefusesTooManyNames(t *testing.T) {
	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)
	body := lookupBody("usernames", maximumLookups+1)

	response := callAPI(router, findUsersByUsernamePath, body, "")

	if response.Code != http.StatusBadRequest {
		t.Errorf("status code: got %d, want %d (invalid_argument)", response.Code, http.StatusBadRequest)
	}
}

func TestFindUsersByIDRefusesAnIDThatIsNotAUUID(t *testing.T) {
	// The fake returns the error the real store returns for a bad id.
	users := &fakeUserStore{err: fmt.Errorf("%w: %q", errInvalidUUID, "x")}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	response := callAPI(router, findUsersByIDPath, `{"userIds": ["x"]}`, "")

	if response.Code != http.StatusBadRequest {
		t.Errorf("status code: got %d, want %d (invalid_argument)", response.Code, http.StatusBadRequest)
	}
}

func TestFindUsersByIDHidesDatabaseErrors(t *testing.T) {
	users := &fakeUserStore{err: errors.New("connection refused")}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)

	response := callAPI(router, findUsersByIDPath, `{"userIds": []}`, "")

	if response.Code != http.StatusInternalServerError {
		t.Errorf("status code: got %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "connection refused") {
		t.Errorf("the database error leaked into the answer: %s", response.Body)
	}
}

// lookupBody builds a JSON body with a list of count made-up values.
func lookupBody(field string, count int) string {
	values := make([]string, 0, count)
	for number := range count {
		value := fmt.Sprintf("user%d", number)
		values = append(values, value)
	}
	body, err := json.Marshal(map[string][]string{field: values})
	if err != nil {
		panic(err) // a map of string lists can always be turned into JSON
	}
	return string(body)
}

// compactJSON removes the spaces Connect may put into its JSON answers, so
// they can be compared as text.
func compactJSON(t *testing.T, text string) string {
	t.Helper()
	var value any
	err := json.Unmarshal([]byte(text), &value)
	if err != nil {
		t.Fatalf("not JSON: %s", text)
	}
	compact, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("could not write JSON: %v", err)
	}
	return string(compact)
}
