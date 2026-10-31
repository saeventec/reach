package main

// The GreetingService API, defined in proto/hello/v1/greeting.proto.
// Learn more: docs/learn/connect-rpc.md

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect/v2"

	hellov1 "github.com/saeventec/reach/services/hello/gen/hello/v1"
)

// greetingStore is what the greeting server needs from the database.
// It is an interface, so tests can use a fake store instead of a real
// database. Learn more: docs/learn/testing.md#fakes
type greetingStore interface {
	NewestGreeting(ctx context.Context) (string, error)
}

// greetingServer implements the generated GreetingServiceHandler interface.
type greetingServer struct {
	greetings greetingStore
}

// GetGreeting answers the GetGreeting call with the newest greeting.
func (server *greetingServer) GetGreeting(ctx context.Context, request *hellov1.GetGreetingRequest) (*hellov1.GetGreetingResponse, error) {
	message, err := server.greetings.NewestGreeting(ctx)
	if errors.Is(err, errNoGreeting) {
		return nil, connect.NewError(connect.CodeNotFound, "there is no greeting yet")
	}
	if err != nil {
		// Log the details for us, but send the client only a general
		// message: details can reveal how the system works inside.
		slog.Error("could not read the greeting", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not read the greeting")
	}

	response := &hellov1.GetGreetingResponse{Message: message}
	return response, nil
}
