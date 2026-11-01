package main

// The AccountService API, defined in proto/auth/v1/account.proto.
// Learn more: docs/learn/connect-rpc.md

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect/v2"

	authv1 "github.com/saeventec/reach/packages/go/gen/auth/v1"
)

// accountServer implements the generated AccountServiceHandler interface.
type accountServer struct {
	users    userStore
	limiters *loginLimiters
}

// SignUp checks the username and the password, hashes the password, and
// saves the new account.
func (server *accountServer) SignUp(ctx context.Context, request *authv1.SignUpRequest) (*authv1.SignUpResponse, error) {
	callInfo, err := serverCallInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Sign-ups count toward the same per-address limit as logins, so one
	// computer can't create accounts in bulk.
	ip := clientIP(callInfo)
	if !server.limiters.byIP.Take(ip) {
		slog.Warn("too many sign-ups from one address", "ip", ip)
		return nil, connect.NewError(connect.CodeResourceExhausted, tooManyAttemptsMessage)
	}

	username := normalizeUsername(request.GetUsername())
	err = validateUsername(username)
	if err != nil {
		// The rule errors are written for people, so the client may see them.
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error())
	}

	password := normalizePassword(request.GetPassword())
	err = validatePassword(password, username)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err.Error())
	}

	passwordHash := hashPassword(password)

	user, err := server.users.CreateUser(ctx, username, passwordHash)
	if errors.Is(err, errUsernameTaken) {
		return nil, connect.NewError(connect.CodeAlreadyExists, err.Error())
	}
	if err != nil {
		// Log the details for us, but send the client only a general
		// message: details can reveal how the system works inside.
		slog.Error("could not create the account", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not create the account")
	}

	// Log the new account's id, but never the password.
	slog.Info("a new account was created", "user_id", user.id)

	response := &authv1.SignUpResponse{
		UserId:   user.id,
		Username: user.username,
	}
	return response, nil
}
