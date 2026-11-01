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

	router := newRouter()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
}
