package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTheAPINeedsTheGatewaysUserHeader(t *testing.T) {
	// A call without X-User-Id did not pass the gateway's session check.
	// It is refused before the core is used, so the core can be nil here.
	request := httptest.NewRequest(http.MethodPost, "/chat.v1.ConversationService/ListConversations", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := newRouter(nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
