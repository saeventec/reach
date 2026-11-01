package conversation

// The "ports": what the core needs from the outside, as interfaces. The
// core calls these methods without knowing who implements them: the
// postgres package in the real service, small fakes in the tests.
// Learn more: docs/learn/hexagonal-architecture.md#ports-and-adapters

import "context"

// Store keeps conversations and their members.
type Store interface {
	// The conversations the store returns have their members' ids, but no
	// usernames: the store doesn't know them.

	// FindOrCreateDirectConversation returns the 1:1 conversation with
	// this pair key, and creates it, with both members, if it doesn't
	// exist yet.
	FindOrCreateDirectConversation(ctx context.Context, pairKey string, firstUserID string, secondUserID string) (Conversation, error)

	// CreateGroupConversation creates a group with these members.
	CreateGroupConversation(ctx context.Context, name string, memberIDs []string) (Conversation, error)

	// ListConversationsOfUser returns the user's conversations, the most
	// recently active first.
	ListConversationsOfUser(ctx context.Context, userID string) ([]Conversation, error)
}

// UserDirectory finds people by username or id. The auth service owns
// the accounts, so chat asks it.
type UserDirectory interface {
	// FindUsersByUsername returns the users with these usernames. Unknown
	// usernames are left out.
	FindUsersByUsername(ctx context.Context, usernames []string) ([]User, error)

	// FindUsersByID returns the users with these ids. Unknown ids are left
	// out.
	FindUsersByID(ctx context.Context, userIDs []string) ([]User, error)
}
