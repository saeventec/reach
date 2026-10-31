package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// fakeSessionStore stands in for the sessions table in tests.
type fakeSessionStore struct {
	err error

	// What CreateSession was asked to save.
	savedUserID    string
	savedTokenHash []byte
	savedExpiresAt time.Time

	// What RenewSession was asked for, and the user it returns.
	renewedTokenHash []byte
	sessionUser      sessionUser
}

func (store *fakeSessionStore) CreateSession(ctx context.Context, userID string, tokenHash []byte, expiresAt time.Time) error {
	if store.err != nil {
		return store.err
	}
	store.savedUserID = userID
	store.savedTokenHash = tokenHash
	store.savedExpiresAt = expiresAt
	return nil
}

func (store *fakeSessionStore) RenewSession(ctx context.Context, tokenHash []byte, newExpiresAt time.Time) (sessionUser, error) {
	if store.err != nil {
		return sessionUser{}, store.err
	}
	store.renewedTokenHash = tokenHash
	return store.sessionUser, nil
}

// testLoginUser is the account the fake user store finds in these tests.
// Its password is "purple tram winter".
var testLoginUser = loginUser{
	id:           testUserID,
	username:     "sadra",
	passwordHash: hashPassword("purple tram winter"),
}

// callAPI sends a Connect call with a JSON body through the router, the
// way the browser does, and returns the recorded response.
func callAPI(router http.Handler, path string, body string, cookieHeader string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	if cookieHeader != "" {
		request.Header.Set("Cookie", cookieHeader)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// findSessionCookie returns the session cookie the response sets, or nil.
func findSessionCookie(recorder *httptest.ResponseRecorder) *http.Cookie {
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == sessionCookieName {
			return cookie
		}
	}
	return nil
}

const (
	logInPath          = "/auth.v1.SessionService/LogIn"
	getCurrentUserPath = "/auth.v1.SessionService/GetCurrentUser"
)

func TestLogInSetsTheSessionCookie(t *testing.T) {
	users := &fakeUserStore{foundUser: testLoginUser}
	sessions := &fakeSessionStore{}
	router := newRouter(users, sessions, true)

	recorder := callAPI(router, logInPath, `{"username": "Sadra", "password": "purple tram winter"}`, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	cookie := findSessionCookie(recorder)
	if cookie == nil {
		t.Fatalf("no session cookie was set")
	}
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("the cookie is not locked down: %v", cookie)
	}
	// The database gets the token's hash, never the token itself.
	wantHash := hashSessionToken(cookie.Value)
	if !bytes.Equal(sessions.savedTokenHash, wantHash) {
		t.Errorf("the saved hash is not the hash of the cookie's token")
	}
	if sessions.savedUserID != testUserID {
		t.Errorf("saved user id: got %q, want %q", sessions.savedUserID, testUserID)
	}
	// The session ends in 30 days. Allow one minute for slow test machines.
	wantExpiresAt := time.Now().Add(sessionLifetime)
	difference := wantExpiresAt.Sub(sessions.savedExpiresAt)
	if difference < 0 || difference > time.Minute {
		t.Errorf("expires at: got %v, want about %v", sessions.savedExpiresAt, wantExpiresAt)
	}
	// The token is only in the cookie, never in the response body.
	if strings.Contains(recorder.Body.String(), cookie.Value) {
		t.Errorf("the response body contains the session token")
	}
}

// A wrong password and an unknown username must look exactly the same.
func TestLogInFailsTheSameWayForAWrongPasswordAndAnUnknownUser(t *testing.T) {
	wrongPasswordUsers := &fakeUserStore{foundUser: testLoginUser}
	unknownUsers := &fakeUserStore{err: errUserNotFound}

	wrongPassword := callAPI(newRouter(wrongPasswordUsers, &fakeSessionStore{}, true), logInPath,
		`{"username": "sadra", "password": "purple tram summer"}`, "")
	unknownUser := callAPI(newRouter(unknownUsers, &fakeSessionStore{}, true), logInPath,
		`{"username": "nobody", "password": "purple tram winter"}`, "")

	if wrongPassword.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: status code: got %d, want %d", wrongPassword.Code, http.StatusUnauthorized)
	}
	if wrongPassword.Body.String() != unknownUser.Body.String() {
		t.Errorf("the answers differ:\nwrong password: %s\nunknown user:   %s", wrongPassword.Body.String(), unknownUser.Body.String())
	}
	if !strings.Contains(wrongPassword.Body.String(), wrongLoginMessage) {
		t.Errorf("body: got %s, want the message %q", wrongPassword.Body.String(), wrongLoginMessage)
	}
	if findSessionCookie(wrongPassword) != nil {
		t.Errorf("a failed login set a session cookie")
	}
}

func TestLogInHidesDatabaseErrors(t *testing.T) {
	users := &fakeUserStore{err: errors.New("connection refused")}
	router := newRouter(users, &fakeSessionStore{}, true)

	recorder := callAPI(router, logInPath, `{"username": "sadra", "password": "purple tram winter"}`, "")

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if strings.Contains(recorder.Body.String(), "connection refused") {
		t.Errorf("the error sent to the client shows internal details: %s", recorder.Body.String())
	}
}

func TestGetCurrentUserReturnsTheUserAndRenewsTheCookie(t *testing.T) {
	sessions := &fakeSessionStore{sessionUser: sessionUser{id: testUserID, username: "sadra"}}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, getCurrentUserPath, `{}`, "theme=dark; reach_session=the-token")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"username":"sadra"`) {
		t.Errorf("body: got %s, want the username sadra", recorder.Body.String())
	}
	wantHash := hashSessionToken("the-token")
	if !bytes.Equal(sessions.renewedTokenHash, wantHash) {
		t.Errorf("the session was not looked up by the hash of the cookie's token")
	}
	cookie := findSessionCookie(recorder)
	if cookie == nil || cookie.Value != "the-token" {
		t.Errorf("the cookie was not sent again: %v", cookie)
	}
}

func TestGetCurrentUserWithoutACookie(t *testing.T) {
	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, true)

	recorder := callAPI(router, getCurrentUserPath, `{}`, "")

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

// An expired or unknown token: the store finds no valid session.
func TestGetCurrentUserWithAnInvalidSession(t *testing.T) {
	sessions := &fakeSessionStore{err: errNoSession}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, getCurrentUserPath, `{}`, "reach_session=an-old-token")

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
	if findSessionCookie(recorder) != nil {
		t.Errorf("an invalid session got a cookie")
	}
}
