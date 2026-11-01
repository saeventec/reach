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

	// FindConversation returns one conversation, with its members and its
	// newest message. The core calls it only after checking that the
	// caller is a member, so the conversation exists.
	FindConversation(ctx context.Context, conversationID string) (Conversation, error)

	// IsMember reports whether the user is a member of the conversation.
	IsMember(ctx context.Context, conversationID string, userID string) (bool, error)

	// SaveMessage stores a message with the conversation's next sequence
	// number, and returns it. If a message with this id exists already, it
	// stores nothing and returns the existing one.
	SaveMessage(ctx context.Context, message NewMessage) (Message, error)

	// ListMessages returns the messages the query asks for, the oldest
	// first.
	ListMessages(ctx context.Context, query MessageQuery) ([]Message, error)

	// MarkDelivered moves the member's delivered watermark up to the
	// sequence number (or to the newest message, if that is lower). If it
	// moves, the store also saves a receipt event. A lower number changes
	// nothing.
	MarkDelivered(ctx context.Context, conversationID string, userID string, sequence int64) error

	// MarkRead does the same for the read watermark, and moves the
	// delivered watermark up with it.
	MarkRead(ctx context.Context, conversationID string, userID string, sequence int64) error
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
