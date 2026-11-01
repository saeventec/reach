// Package api is an adapter: it offers the core (the conversation package)
// to browsers as the chat.v1 Connect API, defined in proto/chat/v1/. It
// turns API messages into the core's types and back, and the core's errors
// into Connect error codes. Learn more: docs/learn/hexagonal-architecture.md
package api

import (
	"context"
	"log/slog"

	"connectrpc.com/connect/v2"
)

// userIDHeader carries the logged-in user's id. Only the gateway sets it,
// after asking auth (ADR 0029).
const userIDHeader = "X-User-Id"

// callerID returns the id of the logged-in user who made this call.
func callerID(ctx context.Context) (string, error) {
	callInfo, found := connect.CallInfoForServerContext(ctx)
	if !found {
		slog.Error("the call has no call info: it was not started by the Connect server")
		return "", connect.NewError(connect.CodeInternal, "could not read the request")
	}

	userID := callInfo.RequestHeader().Get(userIDHeader)
	if userID == "" {
		// The gateway always sets the header for a logged-in user. Without
		// it, the request didn't come through the gateway's session check.
		return "", connect.NewError(connect.CodeUnauthenticated, "you are not logged in")
	}
	return userID, nil
}
