package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiterAllowsTheBurstThenRefuses(t *testing.T) {
	// 3 tries at once, then one per hour: no new token arrives during the test.
	limiter := newRateLimiter(3, time.Hour)

	for try := 1; try <= 3; try++ {
		if !limiter.take("1.2.3.4") {
			t.Fatalf("try %d was refused, want allowed", try)
		}
	}
	if limiter.take("1.2.3.4") {
		t.Errorf("try 4 was allowed, want refused")
	}
	// Each key has its own bucket.
	if !limiter.take("5.6.7.8") {
		t.Errorf("another address was refused")
	}
}

func TestRateLimiterIsEmptyTakesNothing(t *testing.T) {
	limiter := newRateLimiter(1, time.Hour)

	for check := 1; check <= 3; check++ {
		if limiter.isEmpty("sadra") {
			t.Fatalf("check %d: the bucket is empty, but nothing was taken", check)
		}
	}
	limiter.take("sadra")
	if !limiter.isEmpty("sadra") {
		t.Errorf("the bucket is not empty after taking its only token")
	}
}

func TestRemoveFullBucketsKeepsTheUsedOnes(t *testing.T) {
	limiter := newRateLimiter(3, time.Hour)
	limiter.isEmpty("idle") // makes a full bucket
	limiter.take("busy")    // makes a bucket with 2 of 3 tokens left

	limiter.removeFullBuckets()

	_, idleFound := limiter.buckets["idle"]
	_, busyFound := limiter.buckets["busy"]
	if idleFound {
		t.Errorf("the full bucket was kept")
	}
	if !busyFound {
		t.Errorf("the used bucket was removed: its count would be lost")
	}
}

// callAPIFrom sends a Connect call as if it came through the gateway from
// this IP address (in the X-Forwarded-For header).
func callAPIFrom(router http.Handler, path string, body string, ip string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Forwarded-For", ip)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestLogInIsLimitedPerAddress(t *testing.T) {
	users := &fakeUserStore{foundUser: testLoginUser}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(), true)
	// Different usernames, so only the address limit can stop them.
	for try := 1; try <= ipLimitBurst; try++ {
		body := credentialsBody("someone"+strings.Repeat("x", try), "purple tram winter")
		recorder := callAPIFrom(router, logInPath, body, "203.0.113.7")
		if errorCode(t, recorder) == "resource_exhausted" {
			t.Fatalf("try %d was limited, want allowed", try)
		}
	}

	limited := callAPIFrom(router, logInPath, credentialsBody("sadra", "purple tram winter"), "203.0.113.7")
	otherAddress := callAPIFrom(router, logInPath, credentialsBody("sadra", "purple tram winter"), "198.51.100.9")

	if errorCode(t, limited) != "resource_exhausted" {
		t.Errorf("try %d: got %s, want resource_exhausted", ipLimitBurst+1, limited.Body.String())
	}
	if otherAddress.Code != http.StatusOK {
		t.Errorf("another address was limited too: %s", otherAddress.Body.String())
	}
}

func TestLogInIsLimitedPerUsernameAfterFailedTries(t *testing.T) {
	users := &fakeUserStore{foundUser: testLoginUser}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(), true)
	// Each wrong try comes from a different address, as in an attack from
	// many computers, so only the username limit can stop them.
	for try := 1; try <= usernameLimitBurst; try++ {
		address := "203.0.113." + strings.Repeat("1", try)
		recorder := callAPIFrom(router, logInPath, credentialsBody("sadra", "wrong password!"), address)
		if errorCode(t, recorder) != "unauthenticated" {
			t.Fatalf("try %d: got %s, want unauthenticated", try, recorder.Body.String())
		}
	}

	// Now even the right password is refused, until tokens drip back in.
	recorder := callAPIFrom(router, logInPath, credentialsBody("sadra", "purple tram winter"), "198.51.100.9")

	if errorCode(t, recorder) != "resource_exhausted" {
		t.Errorf("got %s, want resource_exhausted", recorder.Body.String())
	}
}

// A person who types the right password is never slowed down by the
// username limit.
func TestSuccessfulLoginsDontCountPerUsername(t *testing.T) {
	users := &fakeUserStore{foundUser: testLoginUser}
	router := newRouter(users, &fakeSessionStore{}, newLoginLimiters(), true)

	for try := 1; try <= usernameLimitBurst+2; try++ {
		recorder := callAPIFrom(router, logInPath, credentialsBody("sadra", "purple tram winter"), "203.0.113.7")
		if recorder.Code != http.StatusOK {
			t.Fatalf("login %d: got %s, want success", try, recorder.Body.String())
		}
	}
}

func TestSignUpIsLimitedPerAddress(t *testing.T) {
	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(), true)
	for try := 1; try <= ipLimitBurst; try++ {
		callAPIFrom(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "203.0.113.7")
	}

	recorder := callAPIFrom(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "203.0.113.7")

	if errorCode(t, recorder) != "resource_exhausted" {
		t.Errorf("got %s, want resource_exhausted", recorder.Body.String())
	}
}

// The address added by our gateway is the last one in X-Forwarded-For.
func TestClientIPUsesTheLastForwardedAddress(t *testing.T) {
	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(), true)
	// Use up the limit of 203.0.113.7.
	for try := 1; try <= ipLimitBurst; try++ {
		callAPIFrom(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "203.0.113.7")
	}

	// A client puts a fake address first. The last one is still 203.0.113.7.
	recorder := callAPIFrom(router, signUpPath, credentialsBody("sadra", "purple tram winter"), "10.0.0.1, 203.0.113.7")

	if errorCode(t, recorder) != "resource_exhausted" {
		t.Errorf("a fake first address got around the limit: %s", recorder.Body.String())
	}
}
