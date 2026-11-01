package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/saeventec/reach/packages/go/ratelimit"
)

func TestTheAPINeedsTheGatewaysUserHeader(t *testing.T) {
	// A call without X-User-Id did not pass the gateway's session check.
	// It is refused before the core is used, so the core can be nil here.
	request := httptest.NewRequest(http.MethodPost, "/chat.v1.ConversationService/ListConversations", strings.NewReader("{}"))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	router := newRouter(nil, nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestSendingTooFastIsRefused(t *testing.T) {
	// A limit with no tokens at all: the very first send is too many. The
	// refusal comes before the core, so the core can be nil here.
	noSendsLeft := ratelimit.New(0, time.Hour)
	body := `{"conversationId":"01a19641-0000-7000-8000-0000000000c1","messageId":"01a19641-0000-7000-8000-000000000001","text":"hi"}`
	request := httptest.NewRequest(http.MethodPost, "/chat.v1.MessageService/SendMessage", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-Id", "01a00000-0000-7000-8000-00000000000a")
	recorder := httptest.NewRecorder()

	router := newRouter(nil, noSendsLeft)
	router.ServeHTTP(recorder, request)

	// Connect answers RESOURCE_EXHAUSTED as HTTP 429.
	if recorder.Code != http.StatusTooManyRequests {
		t.Errorf("status code: got %d, want %d", recorder.Code, http.StatusTooManyRequests)
	}
	if !strings.Contains(recorder.Body.String(), "too fast") {
		t.Errorf("got %s, want the \"too fast\" message", recorder.Body.String())
	}
}
