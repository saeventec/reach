package main

// Tests live next to the code they test, in files ending with _test.go.
// Run them with "task test". Learn more: docs/learn/testing.md

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthAnswersOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if body != "ok" {
		t.Errorf("body: got %q, want %q", body, "ok")
	}
}

func TestHealthRejectsOtherMethods(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	recorder := httptest.NewRecorder()

	router := newRouter(&fakeUserStore{}, &fakeSessionStore{}, newLoginLimiters(defaultIPLimitBurst), true)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusMethodNotAllowed {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}
