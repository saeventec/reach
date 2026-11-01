// Package conversation is the core of the chat service: its types and its
// rules. It knows nothing about Postgres, HTTP or NATS. It only describes,
// with interfaces ("ports", in ports.go), what it needs from the outside.
// The adapters (postgres, api, authclient) implement or call it, and
// main.go connects them (ADR 0030).
// Learn more: docs/learn/hexagonal-architecture.md
package conversation

import "time"

// Kind says whether a conversation is a 1:1 chat or a group.
type Kind string

// The two kinds. Their text is also what the database stores.
const (
	KindDirect Kind = "direct"
	KindGroup  Kind = "group"
)

// Conversation is a 1:1 chat or a group, with its members.
type Conversation struct {
	ID   string
	Kind Kind
	// GroupName is empty for a 1:1 chat.
	GroupName      string
	Members        []Member
	CreatedAt      time.Time
	LastActivityAt time.Time
	// LastMessage is the newest message, or nil if there is none yet. A
	// pointer (*Message) can be nil, meaning "no value"; a plain Message
	// can't.
	LastMessage *Message
}

// Member is one person in a conversation.
type Member struct {
	UserID   string
	Username string
	// The member's watermarks: their devices received the messages up to
	// DeliveredUpTo, and they read them up to ReadUpTo. 0 means none yet.
	DeliveredUpTo int64
	ReadUpTo      int64
}

// Receipt is one member's watermarks in one conversation, after a change.
type Receipt struct {
	ConversationID string
	UserID         string
	DeliveredUpTo  int64
	ReadUpTo       int64
}

// User is a person, as the user directory (the auth service) knows them.
type User struct {
	ID       string
	Username string
}
