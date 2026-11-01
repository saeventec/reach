// Package authclient is an adapter: it implements the core's UserDirectory
// port (conversation.UserDirectory) by calling the auth service's internal
// API with Connect. Learn more: docs/learn/connect-rpc.md#internal-apis
package authclient

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect/v2"
	"connectrpc.com/connect/v2/connecthttp"

	authinternalv1 "github.com/saeventec/reach/packages/go/gen/authinternal/v1"
	"github.com/saeventec/reach/packages/go/gen/authinternal/v1/authinternalv1connect"
	"github.com/saeventec/reach/services/chat/conversation"
)

// callTimeout is the longest we wait for auth. Without a limit, a stuck
// auth service would make every chat request hang too.
const callTimeout = 5 * time.Second

// maximumLookups is the most users auth answers per call (see
// proto/authinternal/v1/user_directory.proto). Longer lists are sent in
// several calls.
const maximumLookups = 100

// Directory finds users through the auth service. It implements
// conversation.UserDirectory.
type Directory struct {
	client authinternalv1connect.UserDirectoryServiceClient
}

// NewDirectory creates a directory that calls the auth service at this
// address, like "http://auth:8080".
func NewDirectory(authURL string) *Directory {
	httpClient := &http.Client{Timeout: callTimeout}
	// A Connect client is built in three steps: the transport (how and
	// where to send), the client, and the generated service client.
	transport := connecthttp.NewTransport(httpClient, authURL)
	connectClient := connect.NewClient(transport)
	client := authinternalv1connect.NewUserDirectoryServiceClient(connectClient)
	return &Directory{client: client}
}

// FindUsersByUsername returns the users with these usernames. Unknown
// usernames are left out.
func (directory *Directory) FindUsersByUsername(ctx context.Context, usernames []string) ([]conversation.User, error) {
	users := []conversation.User{}
	for _, part := range splitIntoParts(usernames) {
		request := &authinternalv1.FindUsersByUsernameRequest{Usernames: part}
		response, err := directory.client.FindUsersByUsername(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("ask auth for users by username: %w", err)
		}
		found := toUsers(response.GetUsers())
		users = append(users, found...)
	}
	return users, nil
}

// FindUsersByID returns the users with these ids. Unknown ids are left
// out.
func (directory *Directory) FindUsersByID(ctx context.Context, userIDs []string) ([]conversation.User, error) {
	users := []conversation.User{}
	for _, part := range splitIntoParts(userIDs) {
		request := &authinternalv1.FindUsersByIdRequest{UserIds: part}
		response, err := directory.client.FindUsersById(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("ask auth for users by id: %w", err)
		}
		found := toUsers(response.GetUsers())
		users = append(users, found...)
	}
	return users, nil
}

// splitIntoParts splits a list into parts of at most maximumLookups
// values, the most auth answers in one call.
func splitIntoParts(values []string) [][]string {
	parts := [][]string{}
	for start := 0; start < len(values); start += maximumLookups {
		end := min(start+maximumLookups, len(values))
		parts = append(parts, values[start:end])
	}
	return parts
}

// toUsers turns the API's User messages into the core's users.
func toUsers(apiUsers []*authinternalv1.User) []conversation.User {
	users := make([]conversation.User, 0, len(apiUsers))
	for _, apiUser := range apiUsers {
		user := conversation.User{ID: apiUser.GetId(), Username: apiUser.GetUsername()}
		users = append(users, user)
	}
	return users
}
