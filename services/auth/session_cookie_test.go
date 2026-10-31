package main

import (
	"net/http"
	"testing"
)

func TestNewSessionCookieIsLockedDown(t *testing.T) {
	cookie := newSessionCookie("the-token", true)

	if cookie.Name != sessionCookieName || cookie.Value != "the-token" {
		t.Errorf("name and value: got %q=%q", cookie.Name, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Errorf("the cookie is not HttpOnly")
	}
	if !cookie.Secure {
		t.Errorf("the cookie is not Secure")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("SameSite: got %v, want Strict", cookie.SameSite)
	}
	if cookie.Path != "/" {
		t.Errorf("path: got %q, want %q", cookie.Path, "/")
	}
	wantMaxAge := 30 * 24 * 60 * 60 // 30 days, in seconds
	if cookie.MaxAge != wantMaxAge {
		t.Errorf("max age: got %d, want %d", cookie.MaxAge, wantMaxAge)
	}
}

func TestNewSessionCookieCanBeNotSecure(t *testing.T) {
	cookie := newSessionCookie("the-token", false)

	if cookie.Secure {
		t.Errorf("the cookie is Secure, but secure was false")
	}
}

func TestReadSessionToken(t *testing.T) {
	testCases := []struct {
		cookieHeader string
		wantToken    string
		wantFound    bool
	}{
		{"reach_session=abc", "abc", true},
		{"theme=dark; reach_session=abc; lang=fa", "abc", true},
		{"theme=dark", "", false},
		{"", "", false},
	}

	for _, testCase := range testCases {
		token, found := readSessionToken(testCase.cookieHeader)
		if token != testCase.wantToken || found != testCase.wantFound {
			t.Errorf("readSessionToken(%q): got (%q, %v), want (%q, %v)",
				testCase.cookieHeader, token, found, testCase.wantToken, testCase.wantFound)
		}
	}
}
