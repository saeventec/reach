package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	savedUserAgent string
	savedExpiresAt time.Time

	// What RenewSession was asked for, and the session it returns.
	renewedTokenHash []byte
	session          currentSession

	// What ListSessions returns.
	listedSessions []storedSession

	// What DeleteSession was asked to delete, and the error it returns.
	deletedSessionID string
	deletedUserID    string
	deleteErr        error
}

func (store *fakeSessionStore) CreateSession(ctx context.Context, userID string, tokenHash []byte, userAgent string, expiresAt time.Time) error {
	if store.err != nil {
		return store.err
	}
	store.savedUserID = userID
	store.savedTokenHash = tokenHash
	store.savedUserAgent = userAgent
	store.savedExpiresAt = expiresAt
	return nil
}

func (store *fakeSessionStore) RenewSession(ctx context.Context, tokenHash []byte, newExpiresAt time.Time) (currentSession, error) {
	if store.err != nil {
		return currentSession{}, store.err
	}
	store.renewedTokenHash = tokenHash
	return store.session, nil
}

func (store *fakeSessionStore) ListSessions(ctx context.Context, userID string) ([]storedSession, error) {
	if store.err != nil {
		return nil, store.err
	}
	return store.listedSessions, nil
}

func (store *fakeSessionStore) DeleteSession(ctx context.Context, sessionID string, userID string) error {
	if store.err != nil {
		return store.err
	}
	store.deletedSessionID = sessionID
	store.deletedUserID = userID
	return store.deleteErr
}

// testLoginUser is the account the fake user store finds in these tests.
// Its password is "purple tram winter".
var testLoginUser = loginUser{
	id:           testUserID,
	username:     "sadra",
	passwordHash: hashPassword("purple tram winter"),
}

// testSession is the session the fake session store finds in these tests.
var testSession = currentSession{
	id:       "019a3f2c-0000-7000-8000-000000000001",
	userID:   testUserID,
	username: "sadra",
}

// callAPI sends a Connect call with a JSON body through the router, the
// way the browser does, and returns the recorded response.
func callAPI(router http.Handler, path string, body string, cookieHeader string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", testUserAgent)
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

// testUserAgent is the browser every test request claims to be.
const testUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:132.0) Gecko/20100101 Firefox/132.0"

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
	if sessions.savedUserAgent != testUserAgent {
		t.Errorf("saved user agent: got %q, want %q", sessions.savedUserAgent, testUserAgent)
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
	sessions := &fakeSessionStore{session: testSession}
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

const (
	listDevicesPath  = "/auth.v1.SessionService/ListDevices"
	logOutPath       = "/auth.v1.SessionService/LogOut"
	logOutDevicePath = "/auth.v1.SessionService/LogOutDevice"
)

// otherDeviceID is the id of a second session of the test user.
const otherDeviceID = "019a3f2c-0000-7000-8000-000000000002"

func TestListDevicesMarksThisDevice(t *testing.T) {
	loginTime := time.Date(2026, 10, 31, 9, 0, 0, 0, time.UTC)
	sessions := &fakeSessionStore{
		session: testSession,
		listedSessions: []storedSession{
			{id: testSession.id, userAgent: testUserAgent, createdAt: loginTime, lastUsedAt: loginTime},
			{id: otherDeviceID, userAgent: "curl/8.5.0", createdAt: loginTime, lastUsedAt: loginTime},
		},
	}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, listDevicesPath, `{}`, "reach_session=the-token")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var responseBody struct {
		Devices []struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			LoggedInAt      string `json:"loggedInAt"`
			IsCurrentDevice bool   `json:"isCurrentDevice"`
		} `json:"devices"`
	}
	err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("the response is not valid JSON: %v", err)
	}
	if len(responseBody.Devices) != 2 {
		t.Fatalf("devices: got %d, want 2", len(responseBody.Devices))
	}
	thisDevice := responseBody.Devices[0]
	otherDevice := responseBody.Devices[1]
	if !thisDevice.IsCurrentDevice || otherDevice.IsCurrentDevice {
		t.Errorf("is current device: got %v and %v, want true and false", thisDevice.IsCurrentDevice, otherDevice.IsCurrentDevice)
	}
	if thisDevice.Name != "Firefox on Linux" {
		t.Errorf("name: got %q, want %q", thisDevice.Name, "Firefox on Linux")
	}
	// In JSON, a Protobuf Timestamp is written as text in the RFC 3339
	// format, always in UTC ("Z").
	if thisDevice.LoggedInAt != "2026-10-31T09:00:00Z" {
		t.Errorf("logged in at: got %q, want %q", thisDevice.LoggedInAt, "2026-10-31T09:00:00Z")
	}
}

func TestListDevicesNeedsALogin(t *testing.T) {
	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, true)

	recorder := callAPI(router, listDevicesPath, `{}`, "")

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestLogOutDeletesTheSessionAndRemovesTheCookie(t *testing.T) {
	sessions := &fakeSessionStore{session: testSession}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, logOutPath, `{}`, "reach_session=the-token")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if sessions.deletedSessionID != testSession.id || sessions.deletedUserID != testUserID {
		t.Errorf("deleted: got (%q, %q), want (%q, %q)", sessions.deletedSessionID, sessions.deletedUserID, testSession.id, testUserID)
	}
	assertSessionCookieRemoved(t, recorder)
}

// Logging out twice, or with an expired session, is not an error: the
// device is logged out either way.
func TestLogOutWithoutAValidSessionStillSucceeds(t *testing.T) {
	sessions := &fakeSessionStore{err: errNoSession}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, logOutPath, `{}`, "reach_session=an-old-token")

	if recorder.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	assertSessionCookieRemoved(t, recorder)
}

// assertSessionCookieRemoved checks that the response tells the browser
// to delete the session cookie.
func assertSessionCookieRemoved(t *testing.T, recorder *httptest.ResponseRecorder) {
	// Helper marks this as a helper function, so a failure is reported at
	// the line of the test that called it.
	t.Helper()

	cookie := findSessionCookie(recorder)
	if cookie == nil {
		t.Fatalf("the response doesn't remove the session cookie")
	}
	// Go reads "Max-Age=0" as MaxAge -1: "delete now".
	if cookie.MaxAge != -1 || cookie.Value != "" {
		t.Errorf("the cookie is not removed: %v", cookie)
	}
}

func TestLogOutDeviceDeletesThatSession(t *testing.T) {
	sessions := &fakeSessionStore{session: testSession}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, logOutDevicePath, `{"deviceId": "`+otherDeviceID+`"}`, "reach_session=the-token")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status code: got %d, want %d (body: %s)", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	// The store is asked for that device, and only among this user's own.
	if sessions.deletedSessionID != otherDeviceID || sessions.deletedUserID != testUserID {
		t.Errorf("deleted: got (%q, %q), want (%q, %q)", sessions.deletedSessionID, sessions.deletedUserID, otherDeviceID, testUserID)
	}
	// Logging out another device leaves this device's cookie alone.
	if findSessionCookie(recorder) != nil {
		t.Errorf("logging out another device changed this device's cookie")
	}
}

func TestLogOutDeviceReportsAnUnknownDevice(t *testing.T) {
	sessions := &fakeSessionStore{session: testSession, deleteErr: errNoSession}
	router := newRouter(&fakeUserStore{}, sessions, true)

	recorder := callAPI(router, logOutDevicePath, `{"deviceId": "not-a-device"}`, "reach_session=the-token")

	if recorder.Code != http.StatusNotFound {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
