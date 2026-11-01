package api

import (
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect/v2"

	"github.com/saeventec/reach/services/chat/conversation"
)

func TestToConnectError(t *testing.T) {
	notFound := &conversation.RequestError{Kind: conversation.ErrNotFound, Message: `nobody has the username "x"`}
	invalid := &conversation.RequestError{Kind: conversation.ErrInvalidRequest, Message: "the group needs a name"}
	// The core wraps errors with fmt.Errorf and %w; they must still be found.
	wrappedNotFound := fmt.Errorf("start: %w", notFound)
	databaseDown := errors.New("connection refused")

	testCases := []struct {
		name        string
		err         error
		wantCode    connect.Code
		wantMessage string
	}{
		{"not found", notFound, connect.CodeNotFound, `nobody has the username "x"`},
		{"invalid", invalid, connect.CodeInvalidArgument, "the group needs a name"},
		{"wrapped", wrappedNotFound, connect.CodeNotFound, `nobody has the username "x"`},
		{"our own failure is hidden", databaseDown, connect.CodeInternal, "could not do it"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := toConnectError(testCase.err, "could not do it")

			var connectError *connect.Error
			if !errors.As(got, &connectError) {
				t.Fatalf("got %v, want a Connect error", got)
			}
			if connectError.Code() != testCase.wantCode || connectError.Message() != testCase.wantMessage {
				t.Errorf("got %v %q, want %v %q", connectError.Code(), connectError.Message(), testCase.wantCode, testCase.wantMessage)
			}
		})
	}
}
