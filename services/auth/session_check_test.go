package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// checkSession sends a session check through the router, the way the
// gateway does, and returns the recorded response.
func checkSession(sessions *fakeSessionStore, cookieHeader string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, "/internal/check-session", nil)
	if cookieHeader != "" {
		request.Header.Set("Cookie", cookieHeader)
	}
	recorder := httptest.NewRecorder()

	router := newRouter(&fakeUserStore{}, sessions, newLoginLimiters(), true)
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCheckSessionAnswersWithTheUserID(t *testing.T) {
	sessions := &fakeSessionStore{session: testSession}

	response := checkSession(sessions, "reach_session=some-token")

	if response.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", response.Code, http.StatusOK)
	}
	userID := response.Header().Get("X-User-Id")
	if userID != testUserID {
		t.Errorf("X-User-Id: got %q, want %q", userID, testUserID)
	}
}

func TestCheckSessionWithoutACookie(t *testing.T) {
	sessions := &fakeSessionStore{session: testSession}

	response := checkSession(sessions, "")

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", response.Code, http.StatusUnauthorized)
	}
	// The browser's Connect client reads this as an "unauthenticated" error.
	wantBody := `{"code":"unauthenticated","message":"you are not logged in"}` + "\n"
	if response.Body.String() != wantBody {
		t.Errorf("body: got %q, want %q", response.Body.String(), wantBody)
	}
	if response.Header().Get("X-User-Id") != "" {
		t.Error("a refused check must not send X-User-Id")
	}
}

func TestCheckSessionWithAnInvalidSession(t *testing.T) {
	// No valid session: unknown or expired (the SQL decides, see findSession).
	sessions := &fakeSessionStore{err: errNoSession}

	response := checkSession(sessions, "reach_session=old-token")

	if response.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", response.Code, http.StatusUnauthorized)
	}
}

func TestCheckSessionHidesDatabaseErrors(t *testing.T) {
	sessions := &fakeSessionStore{err: errors.New("connection refused")}

	response := checkSession(sessions, "reach_session=some-token")

	if response.Code != http.StatusInternalServerError {
		t.Errorf("status code: got %d, want %d", response.Code, http.StatusInternalServerError)
	}
	wantBody := `{"code":"internal","message":"could not check the session"}` + "\n"
	if response.Body.String() != wantBody {
		t.Errorf("body: got %q, want %q", response.Body.String(), wantBody)
	}
}

func TestCheckSessionOnlyAnswersGet(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/internal/check-session", nil)
	recorder := httptest.NewRecorder()

	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(), true)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
