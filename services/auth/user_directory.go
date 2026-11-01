package main

// The UserDirectoryService API, defined in
// proto/authinternal/v1/user_directory.proto: other services turn
// usernames into user ids and back. Browsers can't reach it (see the
// .proto file). Learn more: docs/learn/connect-rpc.md#internal-apis

import (
	"context"
	"errors"
	"log/slog"

	"connectrpc.com/connect/v2"

	authinternalv1 "github.com/saeventec/reach/packages/go/gen/authinternal/v1"
)

// maximumLookups is the most usernames or ids one call may ask for. A
// group's members fit easily. Without a limit, one call could make the
// database search for millions of names.
const maximumLookups = 100

// publicUser is the public part of an account: its id and username.
type publicUser struct {
	id       string
	username string
}

// userDirectoryStore is what the directory needs from the users table.
type userDirectoryStore interface {
	FindUsersByUsernames(ctx context.Context, usernames []string) ([]publicUser, error)
	FindUsersByIDs(ctx context.Context, userIDs []string) ([]publicUser, error)
}

// userDirectoryServer implements the generated UserDirectoryServiceHandler
// interface.
type userDirectoryServer struct {
	users userDirectoryStore
}

// FindUsersByUsername returns the users with these usernames.
func (server *userDirectoryServer) FindUsersByUsername(ctx context.Context, request *authinternalv1.FindUsersByUsernameRequest) (*authinternalv1.FindUsersByUsernameResponse, error) {
	typedUsernames := request.GetUsernames()
	if len(typedUsernames) > maximumLookups {
		return nil, tooManyLookupsError()
	}

	// The same rule as at login: " Sadra " finds the account "sadra".
	usernames := make([]string, 0, len(typedUsernames))
	for _, typed := range typedUsernames {
		username := normalizeUsername(typed)
		usernames = append(usernames, username)
	}

	users, err := server.users.FindUsersByUsernames(ctx, usernames)
	if err != nil {
		slog.Error("could not find users by username", "error", err)
		return nil, connect.NewError(connect.CodeInternal, "could not find the users")
	}

	response := &authinternalv1.FindUsersByUsernameResponse{Users: toAPIUsers(users)}
	return response, nil
}

// FindUsersById returns the users with these ids.
func (server *userDirectoryServer) FindUsersById(ctx context.Context, request *authinternalv1.FindUsersByIdRequest) (*authinternalv1.FindUsersByIdResponse, error) {
	userIDs := request.GetUserIds()
	if len(userIDs) > maximumLookups {
		return nil, tooManyLookupsError()
	}

	users, err := server.users.FindUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, findUsersByIDError(err)
	}

	response := &authinternalv1.FindUsersByIdResponse{Users: toAPIUsers(users)}
	return response, nil
}

// findUsersByIDError turns a store error into an API error: a bad id is
// the caller's mistake, anything else is ours.
func findUsersByIDError(err error) error {
	if errors.Is(err, errInvalidUUID) {
		return connect.NewError(connect.CodeInvalidArgument, "every user id must be a UUID")
	}
	slog.Error("could not find users by id", "error", err)
	return connect.NewError(connect.CodeInternal, "could not find the users")
}

// tooManyLookupsError is the error for a call that asks for too much.
func tooManyLookupsError() error {
	return connect.NewError(connect.CodeInvalidArgument, "at most 100 users per call")
}

// toAPIUsers turns stored users into User messages for the API.
func toAPIUsers(users []publicUser) []*authinternalv1.User {
	apiUsers := make([]*authinternalv1.User, 0, len(users))
	for _, user := range users {
		apiUser := &authinternalv1.User{Id: user.id, Username: user.username}
		apiUsers = append(apiUsers, apiUser)
	}
	return apiUsers
}
