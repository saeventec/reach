package api

import (
	"errors"
	"log/slog"

	"connectrpc.com/connect/v2"

	"github.com/saeventec/reach/services/chat/conversation"
)

// toConnectError turns an error from the core into a Connect error. A
// caller's mistake keeps its message, written for the user. Anything else
// is our problem: it is logged, and the caller only sees failureMessage.
func toConnectError(err error, failureMessage string) error {
	var requestError *conversation.RequestError
	if !errors.As(err, &requestError) {
		slog.Error(failureMessage, "error", err)
		return connect.NewError(connect.CodeInternal, failureMessage)
	}

	code := connect.CodeInvalidArgument
	if errors.Is(requestError, conversation.ErrNotFound) {
		code = connect.CodeNotFound
	}
	return connect.NewError(code, requestError.Message)
}
