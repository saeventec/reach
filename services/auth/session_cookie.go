package main

// The session cookie: how the browser keeps the session token and sends it
// back with every request. Learn more: docs/learn/sessions.md#the-cookie

import (
	"net/http"
	"time"
)

// sessionCookieName is the cookie's name in the browser.
const sessionCookieName = "reach_session"

// sessionLifetime is how long a session lasts after its last use. Each use
// starts the 30 days again (the owner's choice).
const sessionLifetime = 30 * 24 * time.Hour

// newSessionCookie builds the cookie that carries a session token.
func newSessionCookie(token string, secure bool) *http.Cookie {
	cookie := &http.Cookie{
		Name:  sessionCookieName,
		Value: token,
		// Send the cookie with requests to every path of our site.
		Path: "/",
		// How long the browser keeps the cookie, in seconds. It matches the
		// session, and is renewed together with it.
		MaxAge: int(sessionLifetime.Seconds()),
		// The page's JavaScript can't read the cookie, so a script injected
		// into our page can't steal the token (ADR 0022).
		HttpOnly: true,
		// Only send the cookie over HTTPS (from COOKIE_SECURE).
		Secure: secure,
		// Only send the cookie with requests made by our own site, never
		// with requests that another website starts (CSRF protection).
		SameSite: http.SameSiteStrictMode,
	}
	return cookie
}

// newRemovedSessionCookie builds a cookie that tells the browser to delete
// the session cookie. The browser finds the cookie to replace by its name,
// domain and path, so those must match newSessionCookie. The other settings
// are kept the same too, to keep it simple.
func newRemovedSessionCookie(secure bool) *http.Cookie {
	cookie := newSessionCookie("", secure)
	// A negative MaxAge sends "Max-Age=0": delete the cookie now.
	cookie.MaxAge = -1
	return cookie
}

// readSessionToken finds the session token in a request's Cookie header.
// It returns false if there is no session cookie.
func readSessionToken(cookieHeader string) (string, bool) {
	// ParseCookie splits "a=1; reach_session=xyz" into separate cookies.
	cookies, err := http.ParseCookie(cookieHeader)
	if err != nil {
		// A broken Cookie header counts as "not logged in".
		return "", false
	}

	for _, cookie := range cookies {
		if cookie.Name == sessionCookieName {
			return cookie.Value, true
		}
	}
	return "", false
}
