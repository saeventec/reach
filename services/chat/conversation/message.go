package conversation

import "time"

// Message is one message in a conversation.
type Message struct {
	// ID is made by the app that sends the message: a UUIDv7.
	ID             string
	ConversationID string
	// Sequence is the message's place in its conversation: 1, 2, 3 ...
	Sequence int64
	SenderID string
	Text     string
	SentAt   time.Time
}

// NewMessage is what a sender sends, before it is stored.
type NewMessage struct {
	ID             string
	ConversationID string
	SenderID       string
	Text           string
}

// Position says which part of a conversation's messages to list.
type Position int

const (
	// PositionNewest lists the newest messages: to open a conversation.
	PositionNewest Position = iota
	// PositionAfter lists the messages after a sequence number: to catch up.
	PositionAfter
	// PositionBefore lists the messages just before a sequence number: to
	// scroll back.
	PositionBefore
)

// MessageQuery says which messages a store should return.
type MessageQuery struct {
	ConversationID string
	Position       Position
	// Sequence is the number to list after or before. Not used for
	// PositionNewest.
	Sequence int64
	// MaximumCount is the most messages to return.
	MaximumCount int
}

// MessagePage is one page of messages, the oldest first.
type MessagePage struct {
	Messages []Message
	// HasMore is true if there are more messages in the same direction.
	HasMore bool
}
