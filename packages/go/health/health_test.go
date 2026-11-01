package health

// Each service also tests that its router sends GET /healthz here.

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleAnswersOK(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()

	Handle(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if body != "ok" {
		t.Errorf("body: got %q, want %q", body, "ok")
	}
}
